import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route } from 'react-router'
import {hashHistory} from 'react-router';
import injectTapEventPlugin from 'react-tap-event-plugin';
import {Tabs, Tab} from 'material-ui/Tabs';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';

import BidAskTab from './bidask.js';
import OpportunityTab from './opportunity.js';


injectTapEventPlugin();

class App extends React.Component {

    constructor(props) {
        super(props);
        this.state = {value: this.props.location.pathname};
    }

    handleActive(tab) {
        this.setState({value: tab.props.value});
        hashHistory.push(tab.props.value);
    }

    render() {
        var that = this;

        return <MuiThemeProvider>
            <div>
                <Tabs value={this.state.value}>
                    <Tab label={"Bid/Ask"} value={"/bid_ask/btc_usd"} onActive={that.handleActive.bind(that)}></Tab>
                    <Tab label={"Opportunities"} value={"/opportunity/btc_usd"} onActive={that.handleActive.bind(that)}></Tab>
                </Tabs>
                <div id="content">
                    {this.props.children}
                </div>
            </div>
        </MuiThemeProvider>
    }
};

ReactDOM.render((
    <Router history={hashHistory}>
        <Route path="/" component={App}>
            <Route path="bid_ask/:pair" component={BidAskTab} />
            <Route path="opportunity/:pair" component={OpportunityTab} />
        </Route>
    </Router>
), document.getElementById('app'));
