const colors = {
  red: 'rgb(255, 99, 132)',
  orange: 'rgb(255, 159, 64)',
  yellow: 'rgb(255, 205, 86)',
  green: 'rgb(75, 192, 192)',
  blue: 'rgb(54, 162, 235)',
  purple: 'rgb(153, 102, 255)',
  grey: 'rgb(201, 203, 207)'
}

const chart = {
  init() {
    const ctx = document.querySelector("#chart").getContext('2d')
    const data = {
      labels: [],
      datasets: [
        {
          data: [],
          borderColor: colors.red,
          label: 'average',
          fill: false,
          type: 'line',
        },
        {
          data: [],
          label: 'min',
          borderColor: colors.yellow,
          fill: false,
        },
        {
          data: [],
          label: 'max',
          backgroundColor: colors.purple,
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
