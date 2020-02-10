/*
 *    Copyright 2020 bitfly gmbh
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

function drawChart(data, title, target, labelFormatter) {


    var options = {
        chart: {
            type: 'area',
            stacked: false,
            height: 350,
            zoom: {
                type: 'x',
                enabled: true,
                autoScaleYaxis: true
            },
            toolbar: {
                autoSelected: 'zoom'
            },
            animations: {
                enabled: false,
            },
            background: currentTheme === 'light' ? '' : 'rgb(38, 35, 39)' ,
        },
        dataLabels: {
            enabled: false
        },
        markers: {
            size: 0,
        },
        yaxis: {
            labels: {
                formatter: labelFormatter
            },
            title: {
                text: title
            },
        },
        series: data,
        xaxis: {
            type: 'datetime'
        },
        title: {
            text: title,
            align: "left"
        },
        tooltip: {
            shared: false,
            y: {
                formatter: labelFormatter
            }
        },
        theme: {
            mode: currentTheme === 'light' ? 'light' : 'dark',
            palette: 'palette6'
        }
    }

    var chart = new ApexCharts(document.querySelector(target), options);

    chart.render();

    return chart;
}