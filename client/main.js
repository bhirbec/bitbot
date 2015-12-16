(function() {

var BidAskTab = require('./bidask.js'),
    OpportunityTab = require('./opportunity.js');

var Router = function (path) {
    var content = document.getElementById('content');

    if (stringStartsWith(path, '/bid_ask' )) {
        $.get(path, function (data) {
            ReactDOM.render(<BidAskTab data={data} />, content);
        });
    } else if (stringStartsWith(path, '/opportunity')) {
        $.get(path, function (data) {
            ReactDOM.render(<OpportunityTab data={data} />, content);
        });
    }
    else {
        content.innerHTML = 'Page not found...'
    }
};

var App = React.createClass({
    render: function () {
        return <div>
            <Tabs />
            <div id="content"></div>
        </div>
    }
});

var Tabs = React.createClass({
    render: function () {
        return <ul>
            <li><a href="#/bid_ask">Bid/Ask</a></li>
            <li><a href="#/opportunity">Opportunities</a></li>
        </ul>
    }
});

function stringStartsWith(string, prefix) {
    return string.slice(0, prefix.length) == prefix;
}

var getLocationHash = function () {
    var hash = window.location.hash;
    return (hash.length && hash[0] == '#') ? hash.slice(1) : hash;
}

var init = function () {
    ReactDOM.render(<App />, document.getElementById('app'));

    $(window).bind('hashchange', function(e) {
        Router(getLocationHash());
    });

    var hash = getLocationHash();
    if (hash == "") {
        window.location.hash = '/bid_ask';
    } else {
        Router(hash);
    }
}

init();

})();
