var React = require('react');

module.exports =  React.createClass({
    getInitialState: function () {
        return {data: []};
    },

    componentDidMount: function () {
        this._updateState(this.props.location);
    },

    componentWillReceiveProps: function (nextProps) {
        this._updateState(nextProps.location)
    },

    _updateState: function (location) {
        var that = this;
        $.get(location.pathname, location.query, function (data) {
            that.setState({data: data});
        });
    },

    render: function () {
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
