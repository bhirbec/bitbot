import React from 'react';
import ReactDOM from 'react-dom';
var d3 = require('d3');

export default class extends React.Component {

    componentWillReceiveProps(props, state) {
        var el = ReactDOM.findDOMNode(this);
        el.innerHTML = "";
        el.appendChild(createLineChart(props.data));
    }

    render() {
        return <div className="chart" />
    }
};

var margin = {top: 20, right: 20, bottom: 30, left: 50},
    width = 600 - margin.left - margin.right,
    height = 140 - margin.top - margin.bottom;

function createLineChart(data) {
    var formatDate = d3.timeParse("%Y-%m-%d %H:%M");

    for (var i=0; i < data.length; i++) {
        var row = data[i];
        row.date = formatDate(row.Date);
    }

    var svgRoot = document.createElementNS(d3.namespaces.svg, 'svg');

    var svg = d3.select(svgRoot)
        .attr("width", width + margin.left + margin.right)
        .attr("height", height + margin.top + margin.bottom);

    var g = addLineChart(data);
    svgRoot.appendChild(g);
    return svgRoot;
}

function addLineChart(data) {
    // x axis
    var x = d3.scaleTime().range([0, width]);
    x.domain(d3.extent(data, function(d) { return d.date; }));

    var xAxis = d3
        .axisBottom()
        .scale(x)
        .ticks(10, d3.timeMinute)
        .tickFormat(d3.timeFormat("%H:%M"));

    // y axis
    var ymin = d3.min(data, function(d) { return d.BidPrice}),
        ymax = d3.max(data, function(d) { return d.AskPrice}),
        delta = (ymax - ymin) * 0.1;

    var y = d3.scaleLinear()
        .domain([ymin - delta, ymax + delta])
        .range([height, 0]);

    var yAxis = d3
        .axisLeft()
        .scale(y)
        .ticks(4);

    var g = document.createElementNS(d3.namespaces.svg, 'g');

    var container = d3
        .select(g)
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    container.append("g")
        .attr("class", "x axis")
        .attr("transform", "translate(0," + height + ")")
        .call(xAxis);

    container.append("g")
        .attr("class", "y axis")
        .call(yAxis)
        .append("text")
        .attr("transform", "rotate(-90)")
        .attr("y", 6)
        .attr("dy", ".71em")
        .style("text-anchor", "end");
        // .text("Price ($)");

    function addLine(container, attr, color) {
        var line = d3.line()
            .x(function(d) { return x(d.date); })
            .y(function(d) { return y(d[attr]); });

        container.append("path")
            .datum(data)
            .attr("class", "line")
            .attr("d", line)
            .attr("stroke", color);
    }

    addLine(container, 'BidPrice', 'steelblue');
    addLine(container, 'AskPrice', '#FC9E27');
    return g;
}
