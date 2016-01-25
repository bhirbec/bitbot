var React = require('react');

module.exports =  React.createClass({
    render: function () {
        var rows = this.props.data.map(function (r) {
            return <tr>
                <td>{r.StartDate}</td>
                <td>{r.Exchanger}</td>
                <td>{r.Bids[0].Price}</td>
                <td>{r.Asks[0].Price}</td>
            </tr>
        });

        return <table>
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Exchanger</th>
                    <th>Bid</th>
                    <th>Ask</th>
                </tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    }
});
