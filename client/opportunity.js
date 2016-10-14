import React from 'react';
import ReactDOM from 'react-dom';
import {hashHistory} from 'react-router';

import SelectField from 'material-ui/SelectField';
import MenuItem from 'material-ui/MenuItem';

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
        var minProfit = form.min_profit.value;
        var limit = form.limit.value;
        hashHistory.push('/opportunity/' + pair + '?min_profit=' + minProfit + '&limit=' + limit);
    }

    render() {
        return <form onSubmit={this.handleSubmit.bind(this)}>
            <SelectField value={this.props.pair} onChange={this.handleChange.bind(this)}>
                {pairs.map(function (p) {
                    return <MenuItem value={p.symbol} primaryText={p.label} />
                })}
            </SelectField>
            <label>Min profit</label>
            <input name="min_profit" type="text" size="10" defaultValue={this.props.location.query.min_profit} />
            <label>Limit</label>
            <input name="limit" type='text' size="10" defaultValue={this.props.location.query.limit} />
            {/* TODO: onSubmit isn't triggered whithout if the form doesn't contain that button.
            I don't understand why... */}
            <input type="submit" value="send" />
        </form>
    }
};

class Table extends React.Component {

    render() {
        if (this.props.data.length == 0) {
            return <p>No results.</p>
        }

        var rows = this.props.data.map(function (r) {
            return <tr>
                <td>{r.Date}</td>
                <td>{r.Spread}%</td>
                <td>{r.Volume}</td>
                <td>{r.BuyExchanger}</td>
                <td>{r.BuyPrice}</td>
                <td>{r.SellExchanger}</td>
                <td>{r.SellPrice}</td>
            </tr>
        });

        return <table>
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Spread</th>
                    <td>Volume</td>
                    <th colSpan="2">Buy</th>
                    <th colSpan="2">Sell</th>
                </tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    }
};
