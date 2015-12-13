var BidAskTable = React.createClass({
    getInitialState: function() {
        return {data: []};
    },
    componentDidMount: function () {
        var that = this
        $.get('/bid_ask', function (data) {
            that.setState({data: data});
        });
    },
    render: function () {
        if (this.state.data.length == 0) {
            return null;
        }
        var rows = this.state.data.map(function (r) {
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

ReactDOM.render(<BidAskTable />, document.getElementById('content'));
