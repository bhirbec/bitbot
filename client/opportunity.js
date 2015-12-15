module.exports = React.createClass({
    render: function () {
        if (this.props.data.length == 0) {
            return <p>No results.</p>
        }

        var rows = this.props.data.map(function (r) {
            return <tr>
                <td>{r.Date}</td>
                <td>{r.Spread}%</td>
                <td>{r.BuyExchanger}</td>
                <td>{r.Ask.Price}</td>
                <td>{r.Ask.Volume}</td>
                <td>{r.SellExchanger}</td>
                <td>{r.Bid.Price}</td>
                <td>{r.Bid.Volume}</td>
            </tr>
        });

        return <table>
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Spread</th>
                    <th colSpan="3">Buy</th>
                    <th colSpan="3">Sell</th>
                </tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    }
});
