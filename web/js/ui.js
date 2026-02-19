/**
 * ui.js — Display helper functions and forecast rendering.
 *
 * Pure UI utilities with no side-effects on load.
 * Depends on shared globals from app.js (selectedSite, siteForecasts, SITES, map).
 */

/**
 * Convert a wind direction in degrees to a 16-point compass string.
 * @param {number} degrees - Wind direction (0–360).
 * @returns {string} Compass direction (e.g. "NNW").
 */
function compassDir(degrees) {
  var directions = [
    'N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE',
    'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'
  ];
  return directions[Math.round(degrees / 22.5) % 16];
}

/**
 * Return a string of star emoji for a 1–5 score.
 * @param {number} count - Number of stars.
 * @returns {string} Star emoji string.
 */
function starsHTML(count) {
  return '⭐'.repeat(count);
}

/**
 * Return a CSS class name for the wind gradient severity.
 * @param {string} gradient - Gradient label (e.g. "Low", "Medium", "High").
 * @returns {string} CSS class name.
 */
function gradientClass(gradient) {
  if (gradient.includes('Low')) return 'gradient-low';
  if (gradient.includes('Medium')) return 'gradient-med';
  return 'gradient-high';
}

/**
 * Return an icon for the wind gradient severity.
 * @param {string} gradient - Gradient label.
 * @returns {string} Emoji icon.
 */
function gradientIcon(gradient) {
  if (gradient.includes('Low')) return '✅';
  if (gradient.includes('Medium')) return '⚠️';
  return '🔴';
}

/**
 * Return an icon for the thermal rating.
 * @param {string} rating - Thermal rating string.
 * @returns {string} Emoji icon.
 */
function thermalIcon(rating) {
  var icons = {
    None: '❄️',
    Weak: '🌤',
    Moderate: '☀️',
    Strong: '🔥',
    Extreme: '⚡'
  };
  return icons[rating] || '❓';
}

/**
 * Return a cloud cover icon based on percentage.
 * @param {number} cover - Cloud cover percentage (0–100).
 * @returns {string} Emoji icon.
 */
function cloudIcon(cover) {
  if (cover < 20) return '☀️';
  if (cover < 50) return '⛅';
  if (cover < 80) return '🌥';
  return '☁️';
}

/**
 * Format precipitation and probability into a display string.
 * @param {number} precipitation - Precipitation amount in mm.
 * @param {number} probability - Precipitation probability percentage.
 * @returns {string} HTML string for rain display.
 */
function rainStr(precipitation, probability) {
  if (precipitation > 0) {
    return '<span class="rain">🌧' + precipitation.toFixed(1) + '</span>';
  }
  if (probability > 30) {
    return probability.toFixed(0) + '%';
  }
  return '-';
}

/**
 * Escape a string for safe insertion into HTML.
 * @param {string} str - Raw string.
 * @returns {string} HTML-escaped string.
 */
function escHtml(str) {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

/**
 * Render the site list in the sidebar.
 * Reads from SITES, siteForecasts, and selectedSite globals.
 */
function renderSiteList() {
  var html = SITES.map(function (site) {
    var forecast = siteForecasts[site.name];
    var escapedName = escHtml(site.name);
    var activeClass = (selectedSite === site.name) ? 'active' : '';
    var scoreHtml = forecast ? starsHTML(forecast.bestScore) : '';

    return '<div class="site-item ' + activeClass + '" data-site="' + escapedName + '">' +
      '<div>' +
        '<div class="site-name">' + escapedName + '</div>' +
        '<div class="site-meta">' + compassDir(site.aspect) + ' facing · ' + site.elevation + 'm</div>' +
      '</div>' +
      '<div class="site-score">' + scoreHtml + '</div>' +
    '</div>';
  }).join('');

  var siteListEl = document.getElementById('siteList');
  siteListEl.innerHTML = html;

  // Event delegation for site clicks (avoids inline onclick with site names)
  siteListEl.onclick = function (e) {
    var item = e.target.closest('.site-item');
    if (item && item.dataset.site) {
      selectSite(item.dataset.site);
    }
  };
}

/**
 * Select a site: highlight it, fly the map to it, and show its forecast.
 * If the forecast is already cached it renders immediately; otherwise
 * it fetches weather data and computes metrics via WASM.
 *
 * @param {string} name - Site name to select.
 */
async function selectSite(name) {
  selectedSite = name;
  renderSiteList();

  var site = SITES.find(function (s) { return s.name === name; });
  if (!site) return;

  map.flyTo([site.lat, site.lon], 12, { duration: 0.5 });
  showForecastPanel();

  var panel = document.getElementById('forecastPanel');

  // Use cached data if available
  if (siteForecasts[name]) {
    renderForecast(siteForecasts[name]);
    return;
  }

  panel.innerHTML =
    '<p style="text-align:center;padding:2rem;"><span class="spinner"></span> Fetching forecast…</p>';
  setStatus('Fetching weather for ' + name + '…');

  try {
    var weatherJSON = await fetchWeather(site);

    if (!wasmReady) {
      panel.innerHTML =
        '<p style="color:var(--bad);text-align:center;padding:2rem;">WASM not loaded</p>';
      return;
    }

    var result = JSON.parse(
      pgforecastWasm.computeMetrics(weatherJSON, JSON.stringify(site), getTuningJSON())
    );

    if (result.error) {
      panel.innerHTML =
        '<p style="color:var(--bad);text-align:center;padding:2rem;">' + result.error + '</p>';
      return;
    }

    var days = groupByDay(result);
    var forecast = {
      site: site,
      days: days,
      bestScore: 0,
      _weatherJSON: weatherJSON
    };

    days.forEach(function (day) {
      day.hours.forEach(function (hour) {
        if (hour.flyability_score > forecast.bestScore) {
          forecast.bestScore = hour.flyability_score;
        }
      });
    });

    siteForecasts[name] = forecast;
    updateMarkerColor(name, forecast.bestScore);
    renderSiteList();
    renderForecast(forecast);
    setStatus(name + ' loaded');
  } catch (err) {
    panel.innerHTML =
      '<p style="color:var(--bad);text-align:center;padding:2rem;">' + err.message + '</p>';
    setStatus('Error: ' + err.message);
  }
}

/**
 * Group an array of hourly metrics by date, keeping only daylight hours.
 * Uses the is_day field from Open-Meteo (based on solar position at the site).
 * @param {Array<Object>} metrics - Hourly metric objects with `time` and `is_day` fields.
 * @returns {Array<{date: string, hours: Array<Object>}>} Grouped days.
 */
function groupByDay(metrics) {
  var dayMap = {};
  var orderedKeys = [];

  metrics.forEach(function (hour) {
    var dt = new Date(hour.time);
    var dateKey = dt.toISOString().slice(0, 10);

    if (!hour.is_day) return;

    if (!dayMap[dateKey]) {
      dayMap[dateKey] = [];
      orderedKeys.push(dateKey);
    }
    dayMap[dateKey].push(hour);
  });

  return orderedKeys.map(function (key) {
    return { date: key, hours: dayMap[key] };
  });
}

/**
 * Render the full forecast panel for a site.
 * Shows detailed hourly tables for the first 3 days and an extended
 * outlook summary for remaining days.
 *
 * @param {Object} forecast - Forecast object with site, days, bestScore.
 */
function renderForecast(forecast) {
  var panel = document.getElementById('forecastPanel');
  var days = forecast.days;
  var detailed = days.slice(0, 3);
  var extended = days.slice(3);

  var html = '<h2 style="margin-bottom:1rem;font-size:1.2rem;">' +
    escHtml(forecast.site.name) + '</h2>';

  detailed.forEach(function (day, index) {
    var dt = new Date(day.date + 'T12:00:00Z');
    var dayString = dt.toLocaleDateString('en-GB', {
      weekday: 'short', day: 'numeric', month: 'short'
    });

    var label;
    if (index === 0) label = 'TODAY';
    else if (index === 1) label = 'TOMORROW';
    else label = dayString;

    var midIndex = Math.floor(day.hours.length / 2);
    var midHour = day.hours[midIndex] || day.hours[0];
    var bestScore = Math.max.apply(null, day.hours.map(function (h) {
      return h.flyability_score;
    }));
    var cloudbase = midHour ? Math.round(midHour.cloudbase_ft || 0) : 0;

    html += '<div class="day-section">' +
      '<div class="day-header"><span class="day-label">' + label + '</span> ' + dayString + '</div>' +
      '<div class="summary-cards">' +
        '<div class="summary-card"><div class="label">Best Score</div><div class="value">' + starsHTML(bestScore) + '</div></div>' +
        '<div class="summary-card"><div class="label">Cloudbase</div><div class="value">' + (cloudbase <= 200 ? 'Fog' : cloudbase + 'ft') + '</div></div>' +
        '<div class="summary-card"><div class="label">CAPE</div><div class="value">' + (midHour ? midHour.cape.toFixed(0) : 0) + ' J/kg</div></div>' +
        '<div class="summary-card"><div class="label">XC Potential</div><div class="value">' + (midHour ? midHour.xc_potential : 'N/A') + '</div></div>' +
      '</div>' +
      '<table class="hour-table">' +
        '<tr><th>Time</th><th>Wind (mph)</th><th>Dir</th><th>Gust (mph)</th>' +
        '<th>Gradient <span class="tooltip-trigger" title="Wind speed difference between 1000hPa (~sea level) and 850hPa (~1500m). High gradient means stronger winds aloft, indicating turbulence risk.">❓</span></th>' +
        '<th>Thermal <span class="tooltip-trigger" title="❄️ None · 🌤 Weak · ☀️ Moderate · 🔥 Strong · ⚡ Extreme. Based on CAPE (convective energy) and lapse rate — higher values mean stronger thermals.">❓</span></th>' +
        '<th>Cloud <span class="tooltip-trigger" title="☀️ &lt;20% · ⛅ 20-50% · 🌥 50-80% · ☁️ &gt;80%. Total cloud cover percentage.">❓</span></th>' +
        '<th>Rain <span class="tooltip-trigger" title="🌧 shows actual precipitation (mm). Percentage shows probability of rain when no precipitation detected.">❓</span></th>' +
        '<th>Score</th></tr>';

    day.hours.forEach(function (h) {
      var t = new Date(h.time);
      var timeStr = t.getUTCHours().toString().padStart(2, '0') + ':00';

      html += '<tr>' +
        '<td>' + timeStr + '</td>' +
        '<td>' + h.wind_speed.toFixed(0) + '</td>' +
        '<td>' + h.wind_dir_str + '</td>' +
        '<td>' + h.wind_gusts.toFixed(0) + '</td>' +
        '<td class="' + gradientClass(h.wind_gradient) + '">' +
          gradientIcon(h.wind_gradient) + ' ' + h.wind_gradient +
          '(+' + h.wind_gradient_diff.toFixed(0) + ')</td>' +
        '<td>' + thermalIcon(h.thermal_rating) + ' ' + h.thermal_rating + '</td>' +
        '<td>' + cloudIcon(h.cloud_cover) + '</td>' +
        '<td>' + rainStr(h.precipitation, h.precip_probability) + '</td>' +
        '<td class="stars">' + starsHTML(h.flyability_score) + '</td>' +
      '</tr>';
    });

    html += '</table></div>';
  });

  // Extended outlook
  if (extended.length > 0) {
    html += '<div class="day-section">' +
      '<div class="day-header"><span class="day-label">EXTENDED OUTLOOK</span></div>' +
      '<table class="extended-table">' +
        '<tr><th>Day</th><th>Wind (mph)</th><th>Dir</th><th>Thermal</th><th>Rain</th><th>Score</th></tr>';

    var thermalRanks = ['None', 'Weak', 'Moderate', 'Strong', 'Extreme'];

    extended.filter(function (day) {
      return day.hours.length > 0;
    }).forEach(function (day) {
      var dt = new Date(day.date + 'T12:00:00Z');
      var dayString = dt.toLocaleDateString('en-GB', {
        weekday: 'short', day: 'numeric', month: 'short'
      });

      var avgWind = day.hours.reduce(function (sum, h) {
        return sum + h.wind_speed;
      }, 0) / day.hours.length;

      var avgDir = day.hours.reduce(function (sum, h) {
        return sum + h.wind_direction;
      }, 0) / day.hours.length;

      var maxPrecipProb = Math.max.apply(null, day.hours.map(function (h) {
        return h.precip_probability;
      }));

      var bestThermalIndex = Math.max.apply(null, day.hours.map(function (h) {
        return thermalRanks.indexOf(h.thermal_rating);
      }));
      var bestThermal = thermalRanks[bestThermalIndex];

      var scores = day.hours.map(function (h) {
        return h.flyability_score;
      }).sort(function (a, b) { return b - a; });

      var topScores = scores.slice(0, 3);
      var avgScore = Math.round(
        topScores.reduce(function (a, b) { return a + b; }, 0) /
        Math.min(3, topScores.length)
      );

      html += '<tr>' +
        '<td>' + dayString + '</td>' +
        '<td>' + avgWind.toFixed(0) + '</td>' +
        '<td>' + compassDir(avgDir) + '</td>' +
        '<td>' + thermalIcon(bestThermal) + ' ' + bestThermal + '</td>' +
        '<td>' + maxPrecipProb.toFixed(0) + '%</td>' +
        '<td class="stars">' + starsHTML(avgScore) + '</td>' +
      '</tr>';
    });

    html += '</table></div>';
  }

  panel.innerHTML = html;
}

/**
 * Update the status bar text at the bottom of the page.
 * @param {string} message - Status message to display.
 */
function setStatus(message) {
  document.getElementById('statusBar').textContent = message;
}
