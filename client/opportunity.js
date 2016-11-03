import React from 'react';
import ReactDOM from 'react-dom';
import {hashHistory} from 'react-router';
import {Table, TableBody, TableHeader, TableHeaderColumn, TableRow, TableRowColumn} from 'material-ui/Table';

import {pairs} from './utils';


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
            <h1>Search for opportunities</h1>
            <SearchForm location={this.props.location} params={this.props.params} />
            <ArbitrageTable data={this.state.data} />
        </div>
    }
};

class SearchForm extends React.Component {

    submit() {
        var form = ReactDOM.findDOMNode(this);
        var qs = $.param({
            minProfit: form.min_profit.value,
            limit: form.limit.value
        })
        hashHistory.push('/opportunity/' + form.pair.value + '?' + qs)
    }

    render() {
        var that = this;

        return <form onSubmit={this.submit.bind(this)} style={ {'float': 'left', 'width': '22em'} }>
            <div className="form-field">
                <label>Pair</label>
                <select name="pair" onChange={this.submit.bind(this)}>
                    {pairs.map(function (p) {
                        return <option value={p.symbol} selected={that.props.params.pair == p.symbol}>{p.label}</option>
                    })}
                </select>
            </div>
            <div className="form-field">
                <label>Min Arbitrage Spread</label>
                <input name="min_profit" type="text" size="10" defaultValue={this.props.location.query.min_profit} />
            </div>
            <div className="form-field">
                <label>Limit</label>
                <input name="limit" type='text' size="10" defaultValue={this.props.location.query.limit} />
            </div>
            {/* TODO: onSubmit isn't triggered whithout if the form doesn't contain that button.
            I don't understand why... */}
            <input type="submit" value="send" />
        </form>
    }
};

class ArbitrageTable extends React.Component {

    render() {
        if (this.props.data.length == 0) {
            return <p>No results.</p>
        }

        var rows = this.props.data.map(function (r) {
            return <TableRow>
                <TableRowColumn>{r.Date}</TableRowColumn>
                <TableRowColumn>{r.Spread}%</TableRowColumn>
                <TableRowColumn>{r.Volume}</TableRowColumn>
                <TableRowColumn>{r.BuyExchanger}</TableRowColumn>
                <TableRowColumn>{r.BuyPrice}</TableRowColumn>
                <TableRowColumn>{r.SellExchanger}</TableRowColumn>
                <TableRowColumn>{r.SellPrice}</TableRowColumn>
            </TableRow>
        });

        return <Table>
            <TableHeader displaySelectAll={false} adjustForCheckbox={false}>
                <TableRow>
                    <TableHeaderColumn>Date</TableHeaderColumn>
                    <TableHeaderColumn>Arbitrage Spread</TableHeaderColumn>
                    <TableHeaderColumn>Volume</TableHeaderColumn>
                    <TableHeaderColumn>Buy Exchanger</TableHeaderColumn>
                    <TableHeaderColumn>Buy Price</TableHeaderColumn>
                    <TableHeaderColumn>Sell Exchanger</TableHeaderColumn>
                    <TableHeaderColumn>Sell Price</TableHeaderColumn>
                </TableRow>
            </TableHeader>
            <TableBody displayRowCheckbox={false}>
                {rows}
            </TableBody>
        </Table>
    }
};
