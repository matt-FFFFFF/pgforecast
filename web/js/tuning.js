/**
 * tuning.js — Scoring parameter tuning system.
 *
 * Provides the tuning overlay UI for adjusting flyability scoring
 * parameters, with import/export and localStorage persistence.
 *
 * Depends on: defaultTuning, activeTuning, TUNING_STORAGE_KEY (app.js),
 *             siteForecasts, SITES, selectedSite (app.js),
 *             setStatus, renderSiteList, renderForecast, groupByDay (ui.js),
 *             fetchWeather (weather.js), updateMarkerColor (map.js).
 */

/**
 * Human-readable labels for each tuning parameter, grouped by section.
 * The `_title` key in each section provides the section heading.
 * @type {Object<string, Object<string, string>>}
 */
var TUNING_LABELS = {
  wind: {
    _title: 'Wind',
    ideal_min: 'Ideal min (mph)',
    ideal_max: 'Ideal max (mph)',
    acceptable_min: 'Acceptable min (mph)',
    acceptable_max: 'Acceptable max (mph)',
    dangerous_max: 'Dangerous max (mph)',
    max_gust_factor: 'Max gust factor',
    dangerous_gust_factor: 'Dangerous gust factor'
  },
  gradient: {
    _title: 'Wind Gradient',
    low_threshold: 'Low threshold (mph diff)',
    high_threshold: 'High threshold (mph diff)',
    high_penalty: 'High penalty',
    medium_penalty: 'Medium penalty'
  },
  thermal: {
    _title: 'Thermals',
    cape_weak: 'CAPE weak (J/kg)',
    cape_moderate: 'CAPE moderate',
    cape_strong: 'CAPE strong',
    cape_extreme: 'CAPE extreme',
    lapse_rate_bonus: 'Lapse rate bonus (°C/km)'
  },
  orographic: {
    _title: 'Orographic Lift',
    min_wind_speed: 'Min wind speed (mph)',
    strong_angle: 'Strong angle (°)',
    moderate_angle: 'Moderate angle (°)',
    weak_angle: 'Weak angle (°)'
  },
  cloudbase: {
    _title: 'Cloudbase',
    min_realistic_ft: 'Min realistic (ft)'
  },
  scoring: {
    _title: 'Flyability Scoring',
    base_score: 'Base score',
    wind_ideal_bonus: 'Wind ideal bonus',
    wind_acceptable_bonus: 'Wind acceptable bonus',
    wind_danger_penalty: 'Wind danger penalty',
    wind_high_penalty: 'Wind high penalty',
    dir_on_bonus: 'Direction on-site bonus',
    dir_off_penalty: 'Direction off-site penalty',
    gust_high_penalty: 'Gust high penalty',
    gust_med_penalty: 'Gust medium penalty',
    rain_penalty: 'Rain penalty',
    rain_prob_penalty: 'Rain probability penalty',
    gradient_high_penalty: 'Gradient high penalty',
    gradient_med_penalty: 'Gradient medium penalty',
    cape_bonus: 'CAPE bonus',
    thermal_strong_bonus: 'Thermal strong bonus'
  },
  xc: {
    _title: 'XC Potential',
    min_cloudbase_ft: 'Min cloudbase (ft)',
    good_cloudbase_ft: 'Good cloudbase (ft)',
    max_wind_speed: 'Max wind speed (mph)',
    min_wind_speed: 'Min wind speed (mph)',
    epic_threshold: 'Epic threshold',
    high_threshold: 'High threshold',
    medium_threshold: 'Medium threshold'
  },
  display: {
    _title: 'Display',
    _nested: true
  }
};

/**
 * Load tuning parameters from tuning.json (or WASM fallback) and
 * restore any user overrides from localStorage.
 */
async function loadTuning() {
  try {
    var response = await fetch('tuning.json');
    defaultTuning = await response.json();
  } catch (e) {
    if (wasmReady) {
      defaultTuning = JSON.parse(pgforecastWasm.defaultTuning());
    }
  }

  if (!defaultTuning) {
    console.error('Failed to load default tuning from tuning.json or WASM');
    return;
  }

  var stored = localStorage.getItem(TUNING_STORAGE_KEY);
  if (stored) {
    try {
      activeTuning = JSON.parse(stored);
      updateTuningBadge(true);
    } catch (e) {
      activeTuning = JSON.parse(JSON.stringify(defaultTuning));
    }
  } else {
    activeTuning = JSON.parse(JSON.stringify(defaultTuning));
  }
}

/**
 * Return the active tuning as a JSON string (for passing to WASM).
 * @returns {string} JSON-encoded tuning object.
 */
function getTuningJSON() {
  return JSON.stringify(activeTuning);
}

/**
 * Check whether the active tuning differs from the defaults.
 * @returns {boolean} True if custom tuning is active.
 */
function hasCustomTuning() {
  return JSON.stringify(activeTuning) !== JSON.stringify(defaultTuning);
}

/**
 * Update the tuning button to show/hide the "CUSTOM" badge.
 * @param {boolean} isCustom - Whether custom tuning is active.
 */
function updateTuningBadge(isCustom) {
  var btn = document.getElementById('tuningBtn');
  if (isCustom) {
    btn.classList.add('active');
    btn.innerHTML = '⚙️ Tuning <span class="tuning-badge">CUSTOM</span>';
  } else {
    btn.classList.remove('active');
    btn.textContent = '⚙️ Tuning';
  }
}

/** Toggle the tuning overlay open/closed. */
function toggleTuning() {
  var overlay = document.getElementById('tuningOverlay');
  if (overlay.classList.contains('open')) {
    closeTuning();
  } else {
    openTuning();
  }
}

/** Open the tuning overlay. */
function openTuning() {
  renderTuningPanel();
  document.getElementById('tuningOverlay').classList.add('open');
}

/** Close the tuning overlay. */
function closeTuning() {
  document.getElementById('tuningOverlay').classList.remove('open');
}

/**
 * Render the display config section (wind strength + gradient colours/icons).
 * @returns {string} HTML string for the display tuning section.
 */
function renderDisplaySection() {
  if (!activeTuning.display) return '';

  var html = '<div class="tuning-section"><h3>Display — Wind Strength</h3><div class="tuning-grid">';
  var ws = activeTuning.display.wind_strength;
  var wsDefault = defaultTuning.display ? defaultTuning.display.wind_strength : {};

  var tiers = ['light', 'moderate', 'fresh', 'strong', 'very_strong'];
  var tierLabels = { light: 'Light', moderate: 'Moderate', fresh: 'Fresh', strong: 'Strong', very_strong: 'Very Strong' };

  tiers.forEach(function (tier) {
    if (!ws[tier]) return;
    var t = ws[tier];
    var d = (wsDefault[tier]) || {};
    var label = tierLabels[tier];

    var rgbChanged = (t.rgb !== d.rgb) ? 'changed' : '';
    var iconChanged = (t.icon !== d.icon) ? 'changed' : '';

    html += '<div class="tuning-field">' +
      '<label>' + label + ' colour</label>' +
      '<div class="color-input-row">' +
        '<input type="color" value="' + t.rgb + '"' +
        ' data-display="wind_strength.' + tier + '.rgb"' +
        ' class="' + rgbChanged + '" onchange="onDisplayInput(this)" />' +
        '<span class="color-hex">' + t.rgb + '</span>' +
      '</div></div>';

    html += '<div class="tuning-field">' +
      '<label>' + label + ' icon</label>' +
      '<input type="text" value="' + t.icon + '"' +
      ' data-display="wind_strength.' + tier + '.icon"' +
      ' class="tuning-text ' + iconChanged + '" onchange="onDisplayInput(this)" />' +
    '</div>';
  });

  html += '</div></div>';

  // Gradient section
  html += '<div class="tuning-section"><h3>Display — Gradient</h3><div class="tuning-grid">';
  var grad = activeTuning.display.gradient;
  var gradDefault = defaultTuning.display ? defaultTuning.display.gradient : {};
  var levels = ['low', 'medium', 'high'];
  var levelLabels = { low: 'Low', medium: 'Medium', high: 'High' };

  levels.forEach(function (level) {
    if (!grad[level]) return;
    var g = grad[level];
    var gd = (gradDefault[level]) || {};

    var rgbChanged = (g.rgb !== gd.rgb) ? 'changed' : '';
    var iconChanged = (g.icon !== gd.icon) ? 'changed' : '';

    html += '<div class="tuning-field">' +
      '<label>' + levelLabels[level] + ' colour</label>' +
      '<div class="color-input-row">' +
        '<input type="color" value="' + g.rgb + '"' +
        ' data-display="gradient.' + level + '.rgb"' +
        ' class="' + rgbChanged + '" onchange="onDisplayInput(this)" />' +
        '<span class="color-hex">' + g.rgb + '</span>' +
      '</div></div>';

    html += '<div class="tuning-field">' +
      '<label>' + levelLabels[level] + ' icon</label>' +
      '<input type="text" value="' + g.icon + '"' +
      ' data-display="gradient.' + level + '.icon"' +
      ' class="tuning-text ' + iconChanged + '" onchange="onDisplayInput(this)" />' +
    '</div>';
  });

  html += '</div></div>';
  return html;
}

/**
 * Handle a display config input change.
 * Updates the active tuning display value.
 * @param {HTMLInputElement} element - The input element that changed.
 */
function onDisplayInput(element) {
  var path = element.dataset.display.split('.');
  var obj = activeTuning.display;
  for (var i = 0; i < path.length - 1; i++) {
    obj = obj[path[i]];
  }
  obj[path[path.length - 1]] = element.value;

  // Update hex label for colour inputs
  if (element.type === 'color') {
    var hex = element.parentElement.querySelector('.color-hex');
    if (hex) hex.textContent = element.value;
  }

  var defaultObj = defaultTuning.display;
  for (var j = 0; j < path.length - 1; j++) {
    defaultObj = defaultObj ? defaultObj[path[j]] : undefined;
  }
  var isChanged = !defaultObj || (element.value !== defaultObj[path[path.length - 1]]);
  element.classList.toggle('changed', isChanged);
}

/**
 * Render the tuning panel HTML with all parameter fields.
 * Each field shows its current value and highlights if changed from default.
 */
function renderTuningPanel() {
  var panel = document.getElementById('tuningPanel');
  var html = '<h2>⚙️ Scoring Parameters <button class="btn-close" onclick="closeTuning()">✕</button></h2>';

  for (var section in TUNING_LABELS) {
    var labels = TUNING_LABELS[section];

    // Handle nested display config separately
    if (labels._nested) {
      html += renderDisplaySection();
      continue;
    }

    html += '<div class="tuning-section"><h3>' + labels._title + '</h3><div class="tuning-grid">';

    for (var key in labels) {
      if (key === '_title') continue;

      var value = activeTuning[section][key];
      var defaultValue = defaultTuning[section][key];
      var changedClass = (value !== defaultValue) ? 'changed' : '';

      html += '<div class="tuning-field">' +
        '<label title="Default: ' + defaultValue + '">' + labels[key] + '</label>' +
        '<input type="number" step="any" value="' + value + '"' +
        ' data-section="' + section + '" data-key="' + key + '"' +
        ' class="' + changedClass + '" onchange="onTuningInput(this)" />' +
      '</div>';
    }

    html += '</div></div>';
  }

  html += '<div class="tuning-actions">' +
    '<button class="btn-apply" onclick="applyTuning()">Apply & Rescore</button>' +
    '<button class="btn-reset" onclick="resetTuning()">Reset to Defaults</button>' +
    '<button class="btn-export" onclick="exportTuning()">📋 Export JSON</button>' +
    '<button class="btn-export" onclick="importTuning()">📂 Import JSON</button>' +
  '</div>';

  panel.innerHTML = html;
}

/**
 * Handle a tuning input field change.
 * Updates the active tuning value and toggles the "changed" highlight.
 *
 * @param {HTMLInputElement} element - The input element that changed.
 */
function onTuningInput(element) {
  var value = parseFloat(element.value);
  if (isNaN(value)) return;

  var section = element.dataset.section;
  var key = element.dataset.key;

  activeTuning[section][key] = value;

  var isChanged = (value !== defaultTuning[section][key]);
  element.classList.toggle('changed', isChanged);
}

/**
 * Save current tuning to localStorage and rescore all sites.
 */
function applyTuning() {
  if (hasCustomTuning()) {
    localStorage.setItem(TUNING_STORAGE_KEY, JSON.stringify(activeTuning));
    updateTuningBadge(true);
  } else {
    localStorage.removeItem(TUNING_STORAGE_KEY);
    updateTuningBadge(false);
  }

  var cachedWeather = {};
  for (var name in siteForecasts) {
    if (siteForecasts[name]._weatherJSON) {
      cachedWeather[name] = siteForecasts[name]._weatherJSON;
    }
  }

  siteForecasts = {};
  rescoreAll(cachedWeather);
  closeTuning();
  setStatus('Tuning applied — rescoring…');
}

/**
 * Reset tuning to defaults, clear localStorage, and rescore all sites.
 */
function resetTuning() {
  activeTuning = JSON.parse(JSON.stringify(defaultTuning));
  localStorage.removeItem(TUNING_STORAGE_KEY);
  updateTuningBadge(false);
  renderTuningPanel();

  var cachedWeather = {};
  for (var name in siteForecasts) {
    if (siteForecasts[name]._weatherJSON) {
      cachedWeather[name] = siteForecasts[name]._weatherJSON;
    }
  }

  siteForecasts = {};
  rescoreAll(cachedWeather);
  setStatus('Reset to defaults — rescored');
}

/**
 * Export the active tuning as a downloadable JSON file.
 */
function exportTuning() {
  var blob = new Blob(
    [JSON.stringify(activeTuning, null, 2)],
    { type: 'application/json' }
  );
  var url = URL.createObjectURL(blob);
  var link = document.createElement('a');
  link.href = url;
  link.download = 'pgforecast-tuning.json';
  link.click();
  setTimeout(function () { URL.revokeObjectURL(url); }, 100);
}

/**
 * Import tuning from a user-selected JSON file.
 * Validates that all required sections are present.
 */
function importTuning() {
  var input = document.createElement('input');
  input.type = 'file';
  input.accept = '.json';

  input.onchange = async function (e) {
    var file = e.target.files[0];
    if (!file) return;

    try {
      var imported = JSON.parse(await file.text());

      // Validate all required sections exist
      for (var section in TUNING_LABELS) {
        if (!imported[section]) {
          throw new Error('Missing: ' + section);
        }
      }

      activeTuning = imported;
      renderTuningPanel();
      setStatus('Imported — click Apply');
    } catch (err) {
      alert('Invalid: ' + err.message);
    }
  };

  input.click();
}

/**
 * Rescore all sites using cached or freshly fetched weather data.
 * Updates the sidebar, map markers, and the active forecast panel.
 *
 * @param {Object<string, string>} cachedWeather - Cached weather JSON keyed by site name.
 */
async function rescoreAll(cachedWeather) {
  for (var i = 0; i < SITES.length; i++) {
    var site = SITES[i];

    try {
      var weatherJSON = cachedWeather[site.name] || await fetchWeather(site);
      var result = JSON.parse(
        pgforecastWasm.computeMetrics(weatherJSON, JSON.stringify(site), getTuningJSON())
      );

      if (!result.error) {
        var metrics = result.metrics || result;
        var days = groupByDay(metrics);
        var bestScore = 0;

        if (result.display && result.wind_thresholds) {
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
      console.error('Failed to rescore ' + site.name + ':', e);
    }

    renderSiteList();
  }

  if (selectedSite && siteForecasts[selectedSite]) {
    renderForecast(siteForecasts[selectedSite]);
  }

  setStatus('All sites rescored');
}
