import React from 'react';
import ReactDOM from 'react-dom';
import {hashHistory} from 'react-router';

import SelectField from 'material-ui/SelectField';
import MenuItem from 'material-ui/MenuItem';
import LineChart from './line-chart';

import {pairs, exchangers} from './utils';


function filterExchangerData(data, exchanger) {
    return data.filter(function (r) {return r.Exchanger == exchanger})
}

export default class extends React.Component {

    constructor(props) {
        super(props);
        this.state = {data: []};
    }

    componentDidMount() {
        this._updateState(this.props.location);
    }

    componentWillReceiveProps(nextProps) {
        this._updateState(nextProps.location)
    }

    _updateState(location) {
        var that = this;
        $.get(location.pathname, location.query, function (data) {
            that.setState({data: data});
        });
    }

    render() {
        return <div>
            <h1>Bid/Ask</h1>
            <SearchForm location={this.props.location} pair={this.props.params.pair} />
            <Table data={this.state.data} />
        </div>
    }
};

class SearchForm extends React.Component {

    handleChange(e, i, pair) {
        this._submit(pair);
        e.preventDefault();
    }

    handleSubmit(e) {
        this._submit(this.props.pair);
        e.preventDefault();
    }

    _submit(pair) {
        var form = ReactDOM.findDOMNode(this);
        hashHistory.push('/bid_ask/' + pair);
    }

    render() {
        return <form onSubmit={this.handleSubmit.bind(this)}>
            <SelectField value={this.props.pair} onChange={this.handleChange.bind(this)}>
                {pairs.map(function (p) {
                    return <MenuItem value={p.symbol} primaryText={p.label} />
                })}
            </SelectField>
            {/* TODO: onSubmit isn't triggered whithout if the form doesn't contain that button.
            I don't understand why... */}
            <input type="submit" value="send" />
        </form>
    }
};

class Table extends React.Component {

    render() {
        var data = this.props.data;

        var rows = exchangers.map(function (ex) {
            var filteredData = filterExchangerData(data, ex)
            var n = filteredData.length;

            return <tr>
                <td>{ex}</td>
                <td>{filteredData[0] ? filteredData[0].BidPrice : '-'}</td>
                <td><LineChart data={filteredData} /></td>
                <td>{filteredData[0] ? filteredData[0].AskPrice : '-'}</td>
            </tr>
        });

        return <table>
            <thead>
                <tr>
                    <th>Exchanger</th>
                    <th>Bid</th>
                    <th>Bid/Ask Evolution</th>
                    <th>Ask</th>
                </tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    }
};
