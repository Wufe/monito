using System;
using Microsoft.AspNetCore.Mvc;
using Monito.Domain.Service.Interface;
using Monito.Web.Models.Input;
using Monito.Web.Services.Interface;

namespace Monito.Web.Controllers.API {

	[ApiController]
	[Route("[controller]")]
	public class JobController : ControllerBase {
		private readonly IJobService _jobService;
		private readonly IHttpRequestService _httpRequestService;
		private readonly IRequestService _requestService;
		private readonly IUserService _userService;

		public JobController(
			IJobService jobService,
			IHttpRequestService httpRequestService,
			IRequestService requestService,
			IUserService userService
		)
		{
			_jobService         = jobService;
			_httpRequestService = httpRequestService;
			_requestService     = requestService;
			_userService        = userService;
		}

		[HttpPost("save")]
		public IActionResult SaveJob([FromBody]SaveJobInputModel inputModel) {
			if (ModelState.IsValid) {
				var user = _userService.FindOrCreateUserByIP(_httpRequestService.GetIP());
				var request = _jobService.BuildRequest(inputModel, user);
				_requestService.AddRequest(request);
				return new JsonResult(new {
					userUUID = user.UUID,
					requestUUID = request.UUID
				});
			} else {
				return BadRequest();
			}
			
		}
	}

}