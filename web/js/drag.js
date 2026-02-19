/**
 * drag.js — Drag handle for resizing the forecast panel.
 *
 * Supports both mouse and touch interactions.
 * Also provides showForecastPanel() and hideForecastPanel().
 *
 * Depends on: map (from app.js).
 */

(function () {
  var handle = document.getElementById('dragHandle');
  var wrapper = document.getElementById('forecastWrapper');
  var main = document.querySelector('.main');

  var dragging = false;
  var startY, startHeight;

  // --- Mouse support ---

  handle.addEventListener('mousedown', function (e) {
    dragging = true;
    startY = e.clientY;
    startHeight = wrapper.offsetHeight;
    wrapper.style.transition = 'none';
    document.body.style.cursor = 'ns-resize';
    document.body.style.userSelect = 'none';
    e.preventDefault();
  });

  document.addEventListener('mousemove', function (e) {
    if (!dragging) return;

    var mainRect = main.getBoundingClientRect();
    var minHeight = 80;
    var maxHeight = mainRect.height - 60;
    var newHeight = Math.max(minHeight, Math.min(maxHeight, startHeight + (startY - e.clientY)));

    wrapper.style.height = newHeight + 'px';
    if (map) map.invalidateSize();
  });

  document.addEventListener('mouseup', function () {
    if (!dragging) return;

    dragging = false;
    wrapper.style.transition = '';
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    if (map) map.invalidateSize();
  });

  // --- Touch support ---

  handle.addEventListener('touchstart', function (e) {
    dragging = true;
    startY = e.touches[0].clientY;
    startHeight = wrapper.offsetHeight;
    wrapper.style.transition = 'none';
    e.preventDefault();
  }, { passive: false });

  document.addEventListener('touchmove', function (e) {
    if (!dragging) return;

    var mainRect = main.getBoundingClientRect();
    var minHeight = 80;
    var maxHeight = mainRect.height - 60;
    var newHeight = Math.max(minHeight, Math.min(maxHeight, startHeight + (startY - e.touches[0].clientY)));

    wrapper.style.height = newHeight + 'px';
    if (map) map.invalidateSize();
  }, { passive: false });

  document.addEventListener('touchend', function () {
    if (!dragging) return;

    dragging = false;
    wrapper.style.transition = '';
    if (map) map.invalidateSize();
  });
})();

/**
 * Show the forecast panel (slide up from bottom).
 * Resets to CSS default height (50%) and invalidates the map after transition.
 */
function showForecastPanel() {
  var wrapper = document.getElementById('forecastWrapper');
  if (!wrapper.classList.contains('open')) {
    wrapper.classList.add('open');
    wrapper.style.height = '';  // reset to CSS default (50%)
    setTimeout(function () {
      if (map) map.invalidateSize();
    }, 300);
  }
}

/**
 * Hide the forecast panel and clear its contents.
 */
function hideForecastPanel() {
  var wrapper = document.getElementById('forecastWrapper');
  wrapper.classList.remove('open');
  wrapper.style.height = '';
  document.getElementById('forecastPanel').innerHTML = '';
  setTimeout(function () {
    if (map) map.invalidateSize();
  }, 300);
}
