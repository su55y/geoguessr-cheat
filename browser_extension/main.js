'use strict'
const localUrl = 'http://localhost:5000/req?u=',
  googleMaps = 'https://google.com/maps/place/'

var lastUrls = []

const getLocationObject = async (u) => {
  const response = await fetch(localUrl + u)
  const data = await response.json()
  return data
}

const printObject = (o) => {
  for (const k of Object.keys(o)) {
    console.log(k, '->', o[k])
    if (k === 'address') printObject(o[k])
  }
}

const guess = (openMap) => {
  if (lastUrls.length > 0) {
    getLocationObject(lastUrls.pop()).then((r) => {
      if (r) {
        console.clear()
        printObject(r)
        if (openMap)
          window.open(`${googleMaps}${r.lat},${r.lon}`, '_blank').focus()
        lastUrls = []
      }
    })
  }
}

const pin5k = () => {
  if (lastUrls.length > 0) {
    getLocationObject(lastUrls.pop()).then((r) => {
      if (r) {
        if (document.querySelector('#method_provider')) {
          document.getElementById('method_provider').remove()
        }
        const t = Date.now()
        // create new element
        let s = document.createElement('script')
        s.id = `method_provider${t}`
        s.innerHTML = `const check${t} = () => {
  let mapObj = document.getElementsByClassName('guess-map__canvas-container')[0] 
  if (!mapObj) return
  mapObj[Object.keys(mapObj).find((key) => key.startsWith('__reactFiber$'))].return.memoizedProps.onMarkerLocationChanged({lat:${r.lat},lng:${r.lon}})
}
check${t}()
document.getElementById("method_provider${t}")?.remove()
`
        document.body.appendChild(s)
        lastUrls = []
      }
    })
  }
}

const isBannedAlready = () => {
  let s = document.getElementById('__NEXT_DATA__')?.innerHTML
  if (!s) return false
  return JSON.parse(s)?.props?.middlewareResults[0]?.account?.isBanned
}

if (isBannedAlready()) {
  alert('btw your account is already banned')
  console.log('account banned:', isBannedAlready())
}

const execute = (e) => {
  switch (e.keyCode) {
    case 67:
      guess(false)
      break
    case 88:
      guess(true)
      break
    case 90:
      pin5k()
      break
  }
}

document.addEventListener('keydown', execute)

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
