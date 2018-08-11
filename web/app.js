const chart = {
  init() {
    const ctx = document.querySelector("#chart").getContext('2d')
    const data = {
      labels: [],
      datasets: [
        {
          data: [],
          label: 'pings',
          backgroundColor: 'rgba(50, 100, 255, 0.2)',
          borderColor: 'rgba(100, 100, 255, 0.5)'
        }
      ]
    }

    this.chart = new Chart(ctx, {
      type: 'line',
      data: data,
    })
  },

  addPings(times) {
    last = this.chart.data.labels[this.chart.data.labels.length-1] || 0
    for (let i = 1; i <= times.length; i++) {
      this.chart.data.labels.push(last + i)
      this.chart.data.datasets[0].data.push(times[i-1])
    }
    this.chart.update()
  },
}

const socket = {
  init() {
    this.socket = new WebSocket(`ws://${location.host}/ws`)
    this.socket.onmessage = this.onmessage.bind(this)
  },
  onmessage(e) {
    chart.addPings([JSON.parse(e.data)])
  }
}

window.addEventListener('load', e => {
  chart.init()
  socket.init()
})
