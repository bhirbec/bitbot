import React from 'react';
import {Table, TableBody, TableHeader, TableHeaderColumn, TableRow, TableRowColumn} from 'material-ui/Table';


export class ArbitrageTab extends React.Component {

    constructor(props) {
        super(props);
        this.state = {data: []};
    }

    componentDidMount() {
        console.log('coucou')
        this._updateState(this.props);
    }

    componentWillReceiveProps(props) {
        this._updateState(props)
    }

    _updateState(props) {
        var that = this;
        $.get("/arbitrage", function (data) {
            that.setState({data: data});
        });
    }

    render() {
        if (this.state.data.length == 0) {
            return <p>No results.</p>
        }

        var rows = this.state.data.map(function (r, i) {
            return <TableRow key={"key-" + i}>
                <TableRowColumn>{r.Date}</TableRowColumn>
                <TableRowColumn>{r.BuyEx}</TableRowColumn>
                <TableRowColumn>{r.SellEx}</TableRowColumn>
                <TableRowColumn>{r.BuyPrice}</TableRowColumn>
                <TableRowColumn>{r.SellPrice}</TableRowColumn>
                <TableRowColumn>{r.Spread} %</TableRowColumn>
                <TableRowColumn>{r.Vol}</TableRowColumn>
            </TableRow>
        });

        return <div>
            <h1>Arbitrage</h1>
            <Table>
                <TableHeader displaySelectAll={false} adjustForCheckbox={false}>
                    <TableRow>
                        <TableHeaderColumn>Date</TableHeaderColumn>
                        <TableHeaderColumn>Buy Exchanger</TableHeaderColumn>
                        <TableHeaderColumn>Sell Exchanger</TableHeaderColumn>
                        <TableHeaderColumn>Buy Price</TableHeaderColumn>
                        <TableHeaderColumn>Sell Price</TableHeaderColumn>
                        <TableHeaderColumn>Arbitrage Spread</TableHeaderColumn>
                        <TableHeaderColumn>Volume</TableHeaderColumn>
                    </TableRow>
                </TableHeader>
                <TableBody displayRowCheckbox={false}>
                    {rows}
                </TableBody>
            </Table>
        </div>
    }
};

export class TradeTab extends React.Component {

    constructor(props) {
        super(props);
        this.state = {data: []};
    }

    componentDidMount() {
        console.log('coucou')
        this._updateState(this.props);
    }

    componentWillReceiveProps(props) {
        this._updateState(props)
    }

    _updateState(props) {
        var that = this;
        $.get("/trade", function (data) {
            that.setState({data: data});
        });
    }

    render() {
        if (this.state.data.length == 0) {
            return <p>No results.</p>
        }

        var rows = this.state.data.map(function (r, i) {
            return <TableRow key={"key-" + i}>
                <TableRowColumn>{r.ArbitrageId}</TableRowColumn>
                <TableRowColumn>{r.TradeId}</TableRowColumn>
                <TableRowColumn>{r.Price}</TableRowColumn>
                <TableRowColumn>{r.Quantity}</TableRowColumn>
                <TableRowColumn>{r.Pair}</TableRowColumn>
                <TableRowColumn>{r.Side}</TableRowColumn>
                <TableRowColumn>{r.Fee}</TableRowColumn>
                <TableRowColumn>{r.FeeCurrency}</TableRowColumn>
            </TableRow>
        });

        return <div>
            <h1>Trades</h1>
            <Table>
                <TableHeader displaySelectAll={false} adjustForCheckbox={false}>
                    <TableRow>
                        <TableHeaderColumn>Arbitrage Id</TableHeaderColumn>
                        <TableHeaderColumn>Trade Id</TableHeaderColumn>
                        <TableHeaderColumn>Price</TableHeaderColumn>
                        <TableHeaderColumn>Quantity</TableHeaderColumn>
                        <TableHeaderColumn>Pair</TableHeaderColumn>
                        <TableHeaderColumn>Side</TableHeaderColumn>
                        <TableHeaderColumn>Fee</TableHeaderColumn>
                        <TableHeaderColumn>Fee Currency</TableHeaderColumn>
                    </TableRow>
                </TableHeader>
                <TableBody displayRowCheckbox={false}>
                    {rows}
                </TableBody>
            </Table>
        </div>
    }
};
