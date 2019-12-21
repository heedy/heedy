async function process(object, data) {
    if (data.length <= 2) {
        return {};
    }
    var isNumber = true;
    if (!data.every(d => !isNaN(d.d))) {
        return {};
    }

    return {
        lineplot: {
            weight: 9,
            title: "Line Plot",
            view: "apexchart",
            data: {
                type: "line",
                series: [{
                    data: data.map(d => [d.t * 1000, d.d])
                }],
                chartOptions: {
                    chart: {
                        animations: {
                            speed: 10,
                            easing: 'linear',
                            dynamicAnimation: {
                                enabled: true
                            },
                        },
                        toolbar: {
                            show: false
                        },
                        zoom: {
                            enabled: false
                        }
                    },
                    dataLabels: {
                        enabled: false
                    },
                    stroke: {
                        curve: 'smooth',
                        width: 3
                    },
                    markers: {
                        size: 3
                    },
                    xaxis: {
                        type: 'datetime'
                    },
                    legend: {
                        show: false
                    },
                    tooltip: {
                        enabled: false
                    }
                }
            }
        }
    };
}

export default process;