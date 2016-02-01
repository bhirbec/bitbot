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
                <Tab label={"Bid/Ask"} value={"/bid_ask/btc_usd"} onActive={that.handleActive}></Tab>
                <Tab label={"Opportunities"} value={"/opportunity/btc_usd"} onActive={that.handleActive}></Tab>
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
