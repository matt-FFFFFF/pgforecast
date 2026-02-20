/**
 * map.js — Leaflet map initialisation and marker management.
 *
 * Depends on: SITES, markers (from app.js), selectSite (from ui.js).
 */

/**
 * Initialise the Leaflet map, add tile layer, and create site markers.
 * Each marker is clickable and shows the site forecast on click.
 * Clicking empty map space deselects the current site.
 */
function initMap() {
  map = L.map('map').setView([50.78, -2.2], 10);

  L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
    attribution: '© OpenStreetMap © CARTO',
    maxZoom: 18
  }).addTo(map);

  SITES.forEach(function (site) {
    var marker = L.circleMarker([site.lat, site.lon], {
      radius: 6,
      fillColor: '#718096',
      color: '#2d3748',
      weight: 2,
      fillOpacity: 0.7
    }).addTo(map);

    marker.bindTooltip(site.name, { direction: 'top', offset: [0, -10] });

    marker.on('click', function (e) {
      L.DomEvent.stopPropagation(e);
      selectSite(site.name);
    });

    markers[site.name] = marker;
  });

  map.on('click', function () {
    if (selectedSite) {
      selectedSite = null;
      clearMarkerHighlight();
      renderSiteList();
      hideForecastPanel();
    }
  });
}

/**
 * Update a marker's colour and size based on the site's best flyability score.
 *
 * @param {string} name - Site name.
 * @param {number} score - Best flyability score (1–5).
 */
function updateMarkerColor(name, score) {
  var colors = { 1: '#f56565', 2: '#ed8936', 3: '#ecc94b', 4: '#48bb78', 5: '#38b2ac' };
  var radii = { 1: 6, 2: 7, 3: 8, 4: 9, 5: 10 };

  var marker = markers[name];
  if (marker) {
    marker.setStyle({
      fillColor: colors[score] || '#718096',
      fillOpacity: score >= 4 ? 0.9 : 0.7
    });
    marker.setRadius(radii[score] || 6);
  }
}

/**
 * Highlight the selected marker with a pulsing ring and reset the previous one.
 * @param {string} name - Site name to highlight.
 */
function highlightMarker(name) {
  // Remove previous highlight
  if (highlightMarker._ring) {
    map.removeLayer(highlightMarker._ring);
    highlightMarker._ring = null;
  }
  if (highlightMarker._prevName && markers[highlightMarker._prevName]) {
    markers[highlightMarker._prevName].setStyle({ weight: 2, color: '#2d3748' });
  }

  var marker = markers[name];
  if (!marker) return;

  // Bold the selected marker border
  marker.setStyle({ weight: 3, color: '#ffffff' });

  // Add a pulsing ring around the selected marker
  var latlng = marker.getLatLng();
  highlightMarker._ring = L.circleMarker(latlng, {
    radius: marker.getRadius() + 8,
    fillColor: 'transparent',
    fillOpacity: 0,
    color: '#ffffff',
    weight: 2,
    opacity: 0.6,
    className: 'marker-pulse'
  }).addTo(map);

  highlightMarker._prevName = name;
}

/**
 * Remove marker highlight (used when deselecting).
 */
function clearMarkerHighlight() {
  if (highlightMarker._ring) {
    map.removeLayer(highlightMarker._ring);
    highlightMarker._ring = null;
  }
  if (highlightMarker._prevName && markers[highlightMarker._prevName]) {
    markers[highlightMarker._prevName].setStyle({ weight: 2, color: '#2d3748' });
    highlightMarker._prevName = null;
  }
}
