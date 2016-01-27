(function() {

var urllib = require('url'),
    React = require('react'),
    ReactDOM = require('react-dom')
    Tabs = require('material-ui/lib/tabs/tabs'),
    Tab = require('material-ui/lib/tabs/tab'),
    injectTapEventPlugin = require('react-tap-event-plugin');

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

// TODO: define list of pairs
var App = React.createClass({
    render: function () {
        return <div>
            <Tabs value={this.state.value}>
                <Tab label="Bid/Ask BTC_USD" value="/bid_ask/btc_usd" onActive={this.handleActive}></Tab>
                <Tab label="Bid/Ask BTC_EUR" value="/bid_ask/btc_eur" onActive={this.handleActive}></Tab>
                <Tab label="Bid/Ask LTC_BTC" value="/bid_ask/ltc_btc" onActive={this.handleActive}></Tab>
                <Tab label="Opportunities BTC_USD" value="/opportunity/btc_usd" onActive={this.handleActive}></Tab>
                <Tab label="Opportunities BTC_EUR" value="/opportunity/btc_eur" onActive={this.handleActive}></Tab>
                <Tab label="Opportunities LTC_BTC" value="/opportunity/ltc_btc" onActive={this.handleActive}></Tab>
            </Tabs>
            <div id="content"></div>
        </div>
    },

    getInitialState: function () {
        return {value: '/bid_ask/btc_usd'};
    },

    handleActive: function (tab) {
        this.setState({value: tab.props.value});
        location.hash = tab.props.value;
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
    injectTapEventPlugin();
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
