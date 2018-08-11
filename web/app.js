const chart = {
  init() {
    const ctx = document.querySelector("#chart").getContext('2d')
    const data = {
      labels: [],
      datasets: [
        {
          data: [],
          backgroundColor: 'rgba(50, 100, 255, 0.2)',
          borderColor: 'rgba(100, 100, 255, 0.8)',
          label: 'average',
          fill: false,
          type: 'line',
        },
        {
          data: [],
          label: 'min',
          backgroundColor: 'rgba(230, 180, 100, 0.5)',
          borderColor: 'rgba(250, 190, 100, 0.8)',
          fill: false,
        },
        {
          data: [],
          label: 'max',
          backgroundColor: 'rgba(255, 81, 58, 0.5)',
          borderColor: 'rgba(255, 96, 76, 0.8)',
          fill: false,
        },
      ]
    }

    this.chart = new Chart(ctx, {
      type: 'bar',
      data: data,
      options: {
        maintainAspectRatio: false,
        title: {
          display: true,
          text: 'Pings',
          position: 'top',
        },
        scales: {
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

  formatTime(timestamp) {
    const date = new Date(timestamp * 1000)
    return `${date.getHours()}:${date.getMinutes()}`
  },

  addStats(statsArr) {
    // { min, max, average, timestamp }
    for (const stats of statsArr) {
      this.chart.data.labels.push(this.formatTime(stats.timestamp))
      this.chart.data.datasets[0].data.push(stats.average)
      this.chart.data.datasets[1].data.push(stats.min)
      this.chart.data.datasets[2].data.push(stats.max)
    }
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
