/**
 * app.js — Shared state declarations and application init.
 *
 * This file loads FIRST so that all shared globals are available
 * to the other modules (ui, weather, map, tuning, drag).
 * The actual init() call happens via <script>init();</script>
 * at the bottom of index.html after every module has loaded.
 */

/** @type {Array<Object>} List of paragliding sites loaded from sites.json */
var SITES = [];

/** @type {Object<string, Object>} Cached forecast data keyed by site name */
var siteForecasts = {};

/** @type {string|null} Currently selected site name */
var selectedSite = null;

/** @type {L.Map|null} Leaflet map instance */
var map = null;

/** @type {Object<string, L.CircleMarker>} Map markers keyed by site name */
var markers = {};

/** @type {boolean} Whether the WASM module has finished loading */
var wasmReady = false;

/** @type {Object|null} Default tuning parameters (from tuning.json or WASM) */
var defaultTuning = null;

/** @type {Object|null} Active tuning parameters (possibly user-customised) */
var activeTuning = null;

/** @type {string} localStorage key for persisted tuning overrides */
var TUNING_STORAGE_KEY = 'pgforecast_tuning';

/**
 * Load all site forecasts in the background.
 * Fetches weather data for each site, computes metrics via WASM,
 * and updates the sidebar and map markers as results arrive.
 */
async function loadAllSitesOverview() {
  setStatus('Loading all sites…');

  for (const site of SITES) {
    try {
      const weatherJSON = await fetchWeather(site);
      const result = JSON.parse(
        pgforecastWasm.computeMetrics(weatherJSON, JSON.stringify(site), getTuningJSON())
      );

      if (!result.error) {
        // WASM output wraps metrics in {metrics, display, wind_thresholds}
        const metrics = result.metrics || result;
        const days = groupByDay(metrics);
        let bestScore = 0;

        // Apply display config from the first successful result
        if (result.display && !displayConfig) {
          displayConfig = result.display;
          windThresholds = result.wind_thresholds;
          applyDisplayConfigCSS(result.display);
        }

        days.slice(0, 3).forEach(function (day) {
          day.hours.forEach(function (hour) {
            if (hour.flyability_score > bestScore) {
              bestScore = hour.flyability_score;
            }
          });
        });

        siteForecasts[site.name] = {
          site: site,
          days: days,
          bestScore: bestScore,
          _weatherJSON: weatherJSON
        };

        updateMarkerColor(site.name, bestScore);
      }
    } catch (e) {
      console.error('Failed to load forecast for ' + site.name + ':', e);
    }

    renderSiteList();
    setStatus('Loaded ' + site.name);
  }

  setStatus('All sites loaded');
}

/**
 * Application entry point.
 * Loads site data, initialises the map, loads WASM + tuning,
 * then kicks off background forecast fetching.
 */
async function init() {
  try {
    SITES = await (await fetch('sites.json')).json();
  } catch (e) {
    setStatus('Failed to load sites.json: ' + e.message);
    return;
  }

  initMap();
  renderSiteList();
  initWindProfilePopups();

  try {
    await loadWasm();
    await loadTuning();
    loadAllSitesOverview();
  } catch (err) {
    document.getElementById('wasmStatus').innerHTML =
      '❌ Failed to load forecast engine';
  }
}
