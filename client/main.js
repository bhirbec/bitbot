(function() {

var Router = function (path) {
    var content = document.getElementById('content');

    if (path == '/bid_ask') {
        $.get('/bid_ask', function (data) {
            ReactDOM.render(<BidAskTable data={data} />, content);
        });
    } else {
        content.innerHTML = 'Page not found.'
    }
};

var App = React.createClass({
    render: function () {
        return <div>
            <Tabs />
            <div id="content"></div>
        </div>
    }
});

var Tabs = React.createClass({
    render: function () {
        return <ul>
            <li><a href="#/bid_ask">Bid/Ask</a></li>
        </ul>
    }
});

var BidAskTable = React.createClass({
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

var getLocationHash = function () {
    var hash = window.location.hash;
    return (hash.length && hash[0] == '#') ? hash.slice(1) : hash;
}

var init = function () {
    ReactDOM.render(<App />, document.getElementById('app'));

    $(window).bind('hashchange', function(e) {
        Router(getLocationHash());
    });

    var hash = getLocationHash();
    if (hash == "") {
        window.location.hash = '/bid_ask';
    } else {
        Router(hash);
    }
}

init();

})();
