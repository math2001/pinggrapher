const chart = {
  init() {
    const ctx = document.querySelector("#chart").getContext('2d')
    const data = {
      labels: [],
      datasets: [
        {
          data: [],
          label: 'average',
          backgroundColor: 'rgba(50, 100, 255, 0.2)',
          borderColor: 'rgba(100, 100, 255, 0.5)'
        },
        {
          data: [],
          label: 'min',
          backgroundColor: 'rgba(255, 150, 100, 0.2)',
          borderColor: 'rgba(255, 150, 100, 0.5)'
        },
        {
          data: [],
          label: 'max',
          backgroundColor: 'rgba(50, 100, 255, 0.2)',
          borderColor: 'rgba(100, 100, 255, 0.5)'
        },
      ]
    }

    this.chart = new Chart(ctx, {
      type: 'line',
      data: data,
      options: {
        maintainAspectRatio: false,
        title: {
          display: true,
          text: 'Pings',
          position: 'top',
        },
        legend: {
          display: false,
        },
        scales: {
          xAxes: [
            {
              ticks: {
                display: false
              },
            }
          ],
          yAxes: [
            {
              ticks: {
                beginAtZero: true,
                callback: label => label + ' ms'
              }
            }
          ]
        }
      }
    })
  },

  addStats(stats) {
    // { min, max, average }
    console.log('add stats', stats)
    last = this.chart.data.labels[this.chart.data.labels.length-1] || 0
    this.chart.data.labels.push(last + 1)
    this.chart.data.datasets[0].data.push(stats.average)
    this.chart.data.datasets[1].data.push(stats.min)
    this.chart.data.datasets[2].data.push(stats.max)
    this.chart.update()
  }
}

const socket = {
  init() {
    this.socket = new WebSocket(`ws://${location.host}/ws`)
    this.socket.onmessage = this.onmessage.bind(this)
  },
  onmessage(e) {
    chart.addStats(JSON.parse(e.data))
  }
}

window.addEventListener('load', e => {
  chart.init()
  socket.init()
})
