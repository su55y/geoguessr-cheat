'use strict'
const localUrl = 'http://localhost:5000/req?u=',
  googleMaps = 'https://google.com/maps/place/'

const KeyBinds = {
  JustPrintToConsole: 67, // c
  PrintAndOpenInGoogleMaps: 88, // x
  PrintAndPinPoint: 90, // z
}

var lastUrls = []

const getLocationObject = async (u) => {
  return await fetch(localUrl + u)
    .then((r) => {
      console.log(`${r.status} ${r.statusText} ${r.url}`)
      if (!r.ok) return false
      return r.json()
    })
    .catch((e) => {
      console.error(e)
      return false
    })
}

const printLocation = (locationObject) => {
  for (const key of Object.keys(locationObject)) {
    console.log(key, '->', locationObject[key])
    if (key === 'address') printLocation(locationObject[key])
  }
}

const openGoogleMaps = ({ lat, lon }) => {
  window.open(`${googleMaps}${lat},${lon}`, '_blank').focus()
}

const guess = (openMap = false) => {
  if (lastUrls.length > 0) {
    getLocationObject(lastUrls.pop()).then((r) => {
      if (!r || typeof r !== 'object') return
      lastUrls.length = 0
      printLocation(r)
      if (openMap) openGoogleMaps(r)
    })
  }
}

const pin5k = () => {
  if (lastUrls.length > 0) {
    getLocationObject(lastUrls.pop()).then((r) => {
      if (!r || typeof r !== 'object') return
      document.getElementById('method_provider')?.remove()
      let s = document.createElement('script')
      s.id = 'method_provider'

      const t = Date.now()
      s.innerHTML = `const pinPoint${t} = () => {
  let mapObj = document.getElementsByClassName('guess-map__canvas-container')[0] 
  if (!mapObj) return
  mapObj[Object.keys(mapObj).find((key) => key.startsWith('__reactFiber$'))].return.memoizedProps.onMarkerLocationChanged({lat:${r.lat},lng:${r.lon}})
}
pinPoint${t}()
document.getElementById('method_provider')?.remove()
`
      document.body.appendChild(s)
      lastUrls.length = 0
      printLocation(r)
    })
  }
}

document.addEventListener('keyup', (e) => {
  switch (e.keyCode) {
    case KeyBinds.JustPrintToConsole:
      guess()
      break
    case KeyBinds.PrintAndOpenInGoogleMaps:
      guess(true)
      break
    case KeyBinds.PrintAndPinPoint:
      pin5k()
      break
  }
})

chrome.runtime.onMessage.addListener((m) => {
  switch (m.message) {
    case 'icon_clicked':
      guess(true)
      break
    case 'request':
      lastUrls.push(m.data.url)
      break
    default:
      console.log('onMessage: unknown message')
  }
})
