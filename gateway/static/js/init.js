// This is needed for CSRF protection used by the backend.
// The token is the one rendered in the <meta> tag above.
(function() {
    var token = document.getElementsByTagName('meta')['gorilla.csrf.Token'].getAttribute('content');
    var oldSend = XMLHttpRequest.prototype.send;
    XMLHttpRequest.prototype.send = function(data) {
        this.setRequestHeader('X-CSRF-Token', token);
        return oldSend.apply(this, arguments);
    };
}());

function openWebsocket(app) {
    // Forward all websocket messages received from our websocket module
    // to the elm runtime where we receive them as subscription.
    // Elm 0.19.0 sadly has no up2date websocket package yet.
    var addr = document.getElementsByTagName('meta')['brig.websocket.addr'].getAttribute('content');
    var ws = new WebSocket(addr);
    ws.onopen = function() {
        console.log("event websocket is open");
    };

    ws.onmessage = function(message) {
        /* console.log(message); */
        app.ports.incoming.send(
            JSON.stringify({
                data: message.data,
                timeStamp: message.timeStamp
            })
        );
    };

    ws.onclose = function(evt) {
        console.log("event websocket was closed");

        // Attempt to reconnect on error.
        // Note that "onclose" is also called when a new connection
        // was attempted but errored (basically always).
        scrollTimer = setTimeout(function() {
            openWebsocket(app)
        }, 2500)
    };

    ws.onerror = function(evt) {
        console.log("event websocket errored: " + evt.data);
        console.log("you might not see updates of your actions.");
    };
}

function pingServer(app) {
    var addr = document.getElementsByTagName('meta')['brig.http.addr'].getAttribute('content');
    var currentState = undefined;

    // Check if the server is still running.
    // If state changes, notify elm app.
    window.setInterval(function() {
        const Http = new XMLHttpRequest();
        Http.open("POST", addr + "/api/v0/ping");
        Http.send();

        Http.onreadystatechange = function() {
            if(this.readyState == 4) {
                if(this.status == 200) {
                    // Only trigger on level change.
                    if(currentState != true) {
                        app.ports.pinger.send(
                            JSON.stringify({
                                isOnline: true,
                            })
                        );
                        currentState = true;
                    }
                } else {
                    // Only trigger on level change.
                    if(currentState != false) {
                        app.ports.pinger.send(
                            JSON.stringify({
                                isOnline: false,
                            })
                        );
                        currentState = false;
                    }
                }
            }
        }
    }, 5000);
}

var processScrollOrResize = function() {
    var _document = window.document;
    var _body = _document.body;
    var _html = _document.documentElement;

    var screenData = {
        scrollTop: parseInt(window.pageYOffset || _html.scrollTop || _body.scrollTop || 0),
        pageHeight: parseInt(
            Math.max(
                _body.scrollHeight,
                _body.offsetHeight,
                _html.clientHeight,
                _html.scrollHeight,
                _html.offsetHeight
            )
        ),
        viewportHeight: parseInt(_html.clientHeight),
        viewportWidth: parseInt(_html.clientWidth),
    };
    app.ports.scrollOrResize.send(screenData);
}

var scrollTimer = null;
var lastScrollFireTime = 0;
var minScrollTime = 150;
var scrolledOrResized = function() {
    if (scrollTimer) {} else {
        var now = new Date().getTime();
        if (now - lastScrollFireTime > minScrollTime) {
            processScrollOrResize();
            lastScrollFireTime = now;
        }
        scrollTimer = setTimeout(function() {
            scrollTimer = null;
            lastScrollFireTime = new Date().getTime();
            processScrollOrResize();
        }, minScrollTime);
    }
};

// Let Elm's runtime take over the "elm" node.
var app = Elm.Main.init({
  node: document.getElementById('elm')
});

app.ports.open.subscribe(function(data) {
    openWebsocket(app);
});

pingServer(app);

window.addEventListener('scroll', scrolledOrResized);
window.addEventListener('resize', scrolledOrResized);
