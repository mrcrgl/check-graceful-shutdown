'use strict';

const http = require('http')
const PHASE_RUNNING = "running";
const PHASE_SHOULD_TERMINATE = "should-terminate";
let phase = PHASE_RUNNING;

// This function will emulate actually doing something to the request by
// replying after some random delay
function handleReq(req, res) {
    let status = 200;
    let body = "OK";

    switch (req.url) {
        case "/health":
            break;
        case "/health/readiness":
            if (phase != PHASE_RUNNING) {
                status = 500;
                body = 'Not so OK';
            }
            break;
    }

    res.writeHead(status);
    res.end(body);

    console.log(`${req.method} ${req.url} - ${status}/${body}`);
}



function startServer(cb) {
    console.log(`Listen on port 8080 with pid=${process.pid}`);
    server.listen(8080, cb);
}

function stopServer() {
    console.log('Stopping server')

    server.close(() => {
        console.log('Server stopped gracefully')
    });
}

process.on('SIGTERM', () => {
    console.info('SIGTERM signal received.');
    console.log('Closing http server in 2s.');
    phase = PHASE_SHOULD_TERMINATE;
    setTimeout(function () {
        console.log('Closing http server now.');
        stopServer();
    }, 2000);
});

const server = http.createServer(handleReq);

startServer(console.log);