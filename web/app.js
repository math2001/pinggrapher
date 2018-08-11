function rand(min, max, len) {
  const arr = []
  for (let i = 0; i <= len; i++) {
    arr.push(Math.random() * (max - min) + min)
  }
  return arr
}

function range(min, max) {
  const arr = []
  for (let i = min; i <= max; i++) {
    arr.push(i)
  }
  return arr
}

const nb = 50

const chart = {
  init() {
    const ctx = document.querySelector("#chart").getContext('2d')
    const data = {
      labels: range(0, nb),
      datasets: [
        {
          data: rand(-nb, nb, nb),
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
    last = this.chart.data.labels[this.chart.data.labels.length-1]
    for (let i = 1; i <= times.length; i++) {
      this.chart.data.labels.push(last + i)
      this.chart.data.datasets[0].data.push(times[i-1])
    }
    this.chart.update()
  },
}

window.addEventListener('load', e => {
  chart.init()
})
