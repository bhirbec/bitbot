var React = require('react'),
    ReactDOM = require('react-dom'),
    Router = require('react-router').Router,
    Route = require('react-router').Route,
    hashHistory = require('react-router').hashHistory;

var Tabs = require('material-ui/lib/tabs/tabs'),
    Tab = require('material-ui/lib/tabs/tab'),
    injectTapEventPlugin = require('react-tap-event-plugin');

var BidAskTab = require('./bidask.js'),
    OpportunityTab = require('./opportunity.js');

var pairs = [
    {symbol: 'btc_usd', label: 'BTC_USD'},
    {symbol: 'btc_eur', label: 'BTC_EUR'},
    {symbol: 'ltc_btc', label: 'LTC_BTC'}
];

injectTapEventPlugin();

var App = React.createClass({
    getInitialState: function () {
        return {value: this.props.location.pathname}
    },

    handleActive: function (tab) {
        this.setState({value: tab.props.value});
        hashHistory.push(tab.props.value);
    },

    render: function () {
        var that = this;

        return <div>
            <Tabs value={this.state.value}>
                {pairs.map(function (p) {
                    return <Tab label={"Bid/Ask " + p.label} value={"/bid_ask/" + p.symbol} onActive={that.handleActive}></Tab>
                })}
                {pairs.map(function (p) {
                    return <Tab label={"Opportunities " + p.label} value={"/opportunity/" + p.symbol} onActive={that.handleActive}></Tab>
                })}
            </Tabs>
            <div id="content">
                {this.props.children}
            </div>
        </div>
    }
});

ReactDOM.render((
    <Router history={hashHistory}>
        <Route path="/" component={App}>
            <Route path="bid_ask/:pair" component={BidAskTab} />
            <Route path="opportunity/:pair" component={OpportunityTab} />
        </Route>
    </Router>
), document.getElementById('app'));
