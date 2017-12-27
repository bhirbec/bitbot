import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import { HashRouter, Link, Route, Redirect, Switch } from 'react-router-dom'
import React from 'react';
import ReactDOM from 'react-dom';
import jquery from 'jquery'

import './main.css'


class App extends React.Component {
    render() {
        return <div>
            <nav className="navbar navbar-default navbar-fixed-top" role="navigation">
                <div className="container">
                    <div className="navbar-header page-scroll">
                        <Link to="/bittrex" className="navbar-brand page-scroll">Bitbot</Link>
                    </div>
                    <div className="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
                        <ul className="nav navbar-nav navbar-right">
                            <li>
                                <Link to="/bittrex" className="page-scroll">Bittrex</Link>
                            </li>
                        </ul>
                    </div>
                </div>
            </nav>
            <div id="main" className="container">
                <Switch>
                    <Route exact path='/' render={() => (<Redirect to='/bittrex' />)} />
                    <Route exact path='/bittrex' component={Bittrex} />
                </Switch>
            </div>
        </div>
    }
}


class Bittrex extends React.Component {
    componentDidMount() {
        jquery.get('http://localhost:8080/api/v1/bittrex').then(resp => {
            this.setState({data: resp || []})
        })
    }

    render() {
        if (this.state == null) {
            return 'Loading...'
        }

        return <div>
            <h1>Bittrex</h1>
            <table className="records">
                <thead>
                    <tr>
                        <th>Market Name</th>
                        <th>Price</th>
                        <th>Volume</th>
                        <th>% 24H</th>
                    </tr>
                </thead>
                <tbody>
                    {this.state.data.map(r => {
                        return <tr key={`price-${r.MarketName}`}>
                            <td>{r.MarketName}</td>
                            <td>{r.Last}</td>
                            <td>{r.Volume}</td>
                            <td>{100 * ((r.Last / r.PrevDay) - 1)}</td>
                        </tr>
                    })}
                </tbody>
            </table>
        </div>
    }
}

ReactDOM.render(<HashRouter><MuiThemeProvider><App /></MuiThemeProvider></HashRouter>, document.getElementById('app'))
