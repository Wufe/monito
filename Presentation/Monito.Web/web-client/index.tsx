import * as React from 'react';
import "core-js/stable";
import "regenerator-runtime/runtime";
import { render } from 'react-dom';
import { App } from '~/app/app';
import { Provider } from 'react-redux';
import { history, store } from '~/state/store';
import { ConnectedRouter } from 'connected-react-router';

const CC = window.cookieconsent;

CC.initialise({
    "palette": {
        "popup": {
            "background": "#2e3440",
            "text": "#eceff4"
        },
        "button": {
            "background": "#bf616a",
            "text": "#eceff4"
        }
    },
    "theme": "edgeless"
});

render(
    <Provider store={store}>
        <ConnectedRouter history={history}>
            <App />
        </ConnectedRouter>
    </Provider>, document.getElementById('app'));