const geoPhotoService =
  'https://maps.googleapis.com/maps/api/js/GeoPhotoService'

chrome.browserAction.onClicked.addListener(() => {
  chrome.tabs.query({ active: true, currentWindow: true }, function (tab) {
    chrome.tabs.sendMessage(tab[0].id, { message: 'icon_clicked' })
  })
})

function logReq(requestDetails) {
  requestDetails?.url.startsWith(geoPhotoService) &&
    chrome.tabs.query({ active: true, currentWindow: true }, function (tab) {
      chrome.tabs.sendMessage(tab[0].id, {
        message: 'request',
        data: requestDetails,
      })
    })
}

chrome.webRequest.onBeforeRequest.addListener(logReq, {
  urls: [`${geoPhotoService}*`],
})
