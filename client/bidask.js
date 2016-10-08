var React = require('react'),
    ReactDOM = require('react-dom'),
    hashHistory = require('react-router').hashHistory;

var SelectField = require('material-ui/lib/select-field'),
    MenuItem = require('material-ui/lib/menus/menu-item');

var pairs = require('./pairs');

module.exports = React.createClass({
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
        return <div>
            <h1>Bid/Ask</h1>
            <SearchForm location={this.props.location} pair={this.props.params.pair} />
            <Table data={this.state.data} />
        </div>
    }
});

var SearchForm = React.createClass({
    handleChange: function (e, i, pair) {
        this._submit(pair);
        e.preventDefault();
    },

    handleSubmit: function (e) {
        this._submit(this.props.pair);
        e.preventDefault();
    },

    _submit: function (pair) {
        var form = ReactDOM.findDOMNode(this);
        hashHistory.push('/bid_ask/' + pair);
    },

    render: function () {
        return <form onSubmit={this.handleSubmit}>
            <SelectField value={this.props.pair} onChange={this.handleChange}>
                {pairs.map(function (p) {
                    return <MenuItem value={p.symbol} primaryText={p.label} />
                })}
            </SelectField>
            {/* TODO: onSubmit isn't triggered whithout if the form doesn't contain that button.
            I don't understand why... */}
            <input type="submit" value="send" />
        </form>
    }
});

var Table =  React.createClass({

    render: function () {
        var rows = this.props.data.map(function (r) {
            return <tr>
                <td>{r.Date}</td>
                <td>{r.Exchanger}</td>
                <td>{r.BidPrice}</td>
                <td>{r.AskPrice}</td>
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
