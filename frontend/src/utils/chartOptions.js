export const baseLineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: { grid: { color: 'rgba(75,85,99,0.2)' } },
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(75,85,99,0.2)' },
      ticks: { color: '#9CA3AF' }
    }
  },
  elements: { point: { radius: 0 } }
}

export const barChartOptions = {
  indexAxis: 'y',
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: {
      beginAtZero: true,
      grid: { color: 'rgba(75,85,99,0.2)' },
      ticks: { color: '#9CA3AF' }
    },
    y: {
      grid: { display: false },
      ticks: { color: '#9CA3AF', font: { size: 11 } }
    }
  }
}

export function createMiniChartOptions(miniChartData) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index',
      intersect: false
    },
    plugins: {
      legend: { 
        display: true,
        labels: { color: '#9CA3AF', font: { size: 10 } }
      }
    },
    scales: {
      x: {
        grid: { color: 'rgba(75,85,99,0.2)' },
        ticks: {
          color: '#9CA3AF',
          maxRotation: 0,
          minRotation: 0,
          autoSkip: true,
          callback: (value) => {
            const raw = miniChartData.value.labels[value] || ''
            const parts = String(raw).split(' ')
            return parts.length > 1 ? parts[1] : raw
          }
        }
      },
      y: {
        position: 'left',
        beginAtZero: true,
        grid: { color: 'rgba(75,85,99,0.15)' },
        ticks: { 
          color: '#9CA3AF', 
          font: { size: 10 },
          stepSize: 1,
          precision: 0
        }
      },
      y1: {
        position: 'right',
        beginAtZero: true,
        grid: { display: false },
        ticks: { 
          color: '#FCA5A5', 
          font: { size: 10 },
          stepSize: 1,
          precision: 0
        }
      }
    },
    elements: {
      point: { radius: 0 }
    }
  }
}
