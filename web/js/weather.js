/**
 * weather.js — Open-Meteo API integration and WASM loader.
 *
 * Globals used: wasmReady (from app.js).
 */

/** @type {number[]} Pressure levels to request from Open-Meteo */
var PRESSURE_LEVELS = [1000, 950, 925, 900, 850, 700];

/** @type {string[]} Surface-level parameters to request */
var SURFACE_PARAMS = [
  'temperature_2m',
  'relative_humidity_2m',
  'dew_point_2m',
  'wind_speed_10m',
  'wind_direction_10m',
  'wind_gusts_10m',
  'cloud_cover',
  'cloud_cover_low',
  'cloud_cover_mid',
  'cloud_cover_high',
  'cape',
  'shortwave_radiation',
  'precipitation',
  'precipitation_probability',
  'freezing_level_height',
  'is_day',
  'weather_code',
  'pressure_msl',
  'visibility'
];

/**
 * Build the Open-Meteo API URL for a given location.
 * Includes both surface parameters and upper-air pressure level data.
 *
 * @param {number} lat - Latitude.
 * @param {number} lon - Longitude.
 * @returns {string} Full API URL.
 */
function buildOpenMeteoURL(lat, lon) {
  var pressureParams = PRESSURE_LEVELS.flatMap(function (p) {
    return [
      'wind_speed_' + p + 'hPa',
      'wind_direction_' + p + 'hPa',
      'temperature_' + p + 'hPa',
      'geopotential_height_' + p + 'hPa'
    ];
  });

  var allParams = SURFACE_PARAMS.concat(pressureParams).join(',');

  return 'https://api.open-meteo.com/v1/forecast' +
    '?latitude=' + lat +
    '&longitude=' + lon +
    '&hourly=' + allParams +
    '&wind_speed_unit=mph' +
    '&forecast_days=16' +
    '&timezone=UTC';
}

/**
 * Fetch weather data from Open-Meteo for a site.
 *
 * @param {Object} site - Site object with lat/lon properties.
 * @returns {Promise<string>} Raw JSON response text.
 * @throws {Error} If the API returns a non-OK status.
 */
async function fetchWeather(site) {
  var response = await fetch(buildOpenMeteoURL(site.lat, site.lon));
  if (!response.ok) {
    throw new Error('API ' + response.status);
  }
  return await response.text();
}

/**
 * Load and initialise the WASM module.
 * Sets wasmReady = true and updates the status indicator on success.
 */
async function loadWasm() {
  var go = new Go();
  var result = await WebAssembly.instantiateStreaming(
    fetch('pgforecast.wasm'),
    go.importObject
  );
  go.run(result.instance);
  wasmReady = true;
  document.getElementById('wasmStatus').innerHTML = '✅ Ready';
}
