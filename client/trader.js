import React from 'react';


export class ArbitrageTab extends React.Component {

    constructor(props) {
        super(props);
        this.state = {data: []};
    }

    componentDidMount() {
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
            return <tr key={"key-" + i}>
                <td>{r.ArbitrageId.substring(0, 10)}...</td>
                <td>{r.Date.substring(0, 16)}</td>
                <td>{r.BuyEx}</td>
                <td>{r.SellEx}</td>
                <td>{r.BuyPrice}</td>
                <td>{r.RealBuyPrice != null ? r.RealBuyPrice.toFixed(6) : '-'}</td>
                <td>{r.SellPrice}</td>
                <td>{r.RealSellPrice != null ? r.RealSellPrice.toFixed(6) : '-'}</td>
                <td>{r.Spread.toFixed(2)}%</td>
                <td>{r.RealSpread != null ? r.RealSpread.toFixed(2) + '%' : '-'}</td>
                <td>{r.Vol}</td>
                <td>{r.RealBuyVol != null ? r.RealBuyVol.toFixed(6) : '-'}</td>
                <td>{r.RealSellVol != null ? r.RealSellVol.toFixed(6) : '-'}</td>
            </tr>
        });

        return <div>
            <h1>Arbitrage</h1>
            <table className="report">
                <thead>
                    <tr>
                        <th>Id</th>
                        <th>Date</th>
                        <th>Buy<br/>Ex</th>
                        <th>Sell<br/>Ex</th>
                        <th>Buy<br/>Price</th>
                        <th>Real Buy<br/>Price</th>
                        <th>Sell<br/>Price</th>
                        <th>Real Sell<br/>Price</th>
                        <th>Margin<br/>(%)</th>
                        <th>Real<br/>Margin (%)</th>
                        <th>Vol</th>
                        <th>Buy Vol</th>
                        <th>Sell Vol</th>
                    </tr>
                </thead>
                <tbody>{rows}</tbody>
            </table>
        </div>
    }
};

export class TradeTab extends React.Component {

    constructor(props) {
        super(props);
        this.state = {data: []};
    }

    componentDidMount() {
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
            return <tr key={"key-" + i}>
                <td>{r.ArbitrageId.substring(5)}...</td>
                <td>{r.TradeId}</td>
                <td>{r.Price}</td>
                <td>{r.Quantity}</td>
                <td>{r.Pair}</td>
                <td>{r.Side}</td>
                <td>{r.Fee}</td>
                <td>{r.FeeCurrency}</td>
            </tr>
        });

        return <div>
            <h1>Trades</h1>
            <table>
                <thead>
                    <tr>
                        <th>Arbitrage Id</th>
                        <th>Trade Id</th>
                        <th>Price</th>
                        <th>Quantity</th>
                        <th>Pair</th>
                        <th>Side</th>
                        <th>Fee</th>
                        <th>Fee Currency</th>
                    </tr>
                </thead>
                <tbody>{rows}</tbody>
            </table>
        </div>
    }
};
