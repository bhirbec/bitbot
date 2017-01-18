import React from 'react';
import ReactDOM from 'react-dom';
import {hashHistory} from 'react-router';
import {Table, TableBody, TableHeader, TableHeaderColumn, TableRow, TableRowColumn} from 'material-ui/Table';

import {pairs, exchangers} from './utils';


export default class extends React.Component {

    constructor(props) {
        super(props);
        this.state = {data: [], params: {}};
    }

    componentDidMount() {
        this._updateState(this.props);
    }

    componentWillReceiveProps(props) {
        this._updateState(props)
    }

    _updateState(props) {
        var that = this;
        $.get(props.location.pathname, props.location.query, function (data) {
            var params = Object.assign({}, props.params, props.location.query);
            that.setState({data: data, params: params});
        });
    }

    render() {
        return <div>
            <h1>Search for opportunities</h1>
            <SearchForm params={this.state.params} />
            <ArbitrageTable data={this.state.data} />
        </div>
    }
};

class SearchForm extends React.Component {
    submit(e) {
        e.preventDefault()

        var form = ReactDOM.findDOMNode(this);

        var qs = $.param({
            buy_ex: form.buy_ex.value,
            sell_ex: form.sell_ex.value,
            min_profit: form.min_profit.value,
            min_vol: form.min_vol.value,
            limit: form.limit.value
        })
        hashHistory.push('/opportunity/' + form.pair.value + '?' + qs)
    }

    componentDidUpdate() {
        var form = ReactDOM.findDOMNode(this);
        for (var key in this.props.params) {
            // TODO: this doesn't work for select
            form[key].value = this.props.params[key]
        }
    }

    render() {
        return <form onSubmit={this.submit.bind(this)} style={ {'float': 'left', 'width': '15em'} }>
            <div className="form-field">
                <label>Pair</label>
                <select name="pair" onChange={this.submit.bind(this)}>
                    {pairs.map(function (p) {
                        return <option key={p.symbol} value={p.symbol}>{p.label}</option>
                    })}
                </select>
            </div>
            <div className="form-field">
                <label>Buy Exchanger</label>
                <select name="buy_ex" onChange={this.submit.bind(this)}>
                    <option value="">All</option>
                    {exchangers.map(function (e) {
                        return <option key={"buy_ex:" + e} value={e}>{e}</option>
                    })}
                </select>
            </div>
            <div className="form-field">
                <label>Sell Exchanger</label>
                <select name="sell_ex" onChange={this.submit.bind(this)}>
                    <option value="">All</option>
                    {exchangers.map(function (e) {
                        return <option key={"sell_ex:" + e} value={e}>{e}</option>
                    })}
                </select>
            </div>
            <div className="form-field">
                <label>Min Arbitrage Spread</label>
                <input name="min_profit" type="text" size="10" />
            </div>
            <div className="form-field">
                <label>Min Volume</label>
                <input name="min_vol" type='text' size="10" />
            </div>
            <div className="form-field">
                <label>Limit</label>
                <input name="limit" type='text' size="10" />
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

        var rows = this.props.data.map(function (r, i) {
            return <TableRow key={"key-" + i}>
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
