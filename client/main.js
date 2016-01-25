(function() {

var urllib = require('url'),
    React = require('react'),
    ReactDOM = require('react-dom');

var BidAskTab = require('./bidask.js'),
    OpportunityTab = require('./opportunity.js');

var Router = function (url) {
    var content = document.getElementById('content');
    var route = urllib.parse(url, true);

    // TODO: don't pass route.pathname. Component should able to modify the has without it.
    if (stringStartsWith(route.pathname , '/bid_ask')) {
        $.get(url, function (data) {
            ReactDOM.render(<BidAskTab uri={route.pathname} data={data} />, content);
        });
    } else if (stringStartsWith(route.pathname, '/opportunity')) {
        $.get(url, function (data) {
            ReactDOM.render(<OpportunityTab uri={route.pathname} data={data} params={route.query} />, content);
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

// TODO: define list of pairs
var Tabs = React.createClass({
    render: function () {
        return <ul>
            <li><a href="#/bid_ask/btc_usd">Bid/Ask BTC_USD</a></li>
            <li><a href="#/bid_ask/btc_eur">Bid/Ask BTC_EUR</a></li>
            <li><a href="#/bid_ask/ltc_btc">Bid/Ask LTC_BTC</a></li>
            <li><a href="#/opportunity/btc_usd">Opportunities BTC_USD</a></li>
            <li><a href="#/opportunity/btc_eur">Opportunities BTC_EUR</a></li>
            <li><a href="#/opportunity/ltc_btc">Opportunities LTC_BTC</a></li>
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
        window.location.hash = '/bid_ask/btc_usd';
    } else {
        Router(hash);
    }
}

init();

})();
