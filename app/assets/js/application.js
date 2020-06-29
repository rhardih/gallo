/** @namespace */
var Gallo = window.Gallo || {};

/**
 * Run-of-the-mill DOM ready ready callback.
 *
 * @see {@link http://youmightnotneedjquery.com/#ready}
 * @param {function} fn - Callback to be invoked on ready state.
 */
Gallo.ready = function(fn) {
  if (document.readyState != 'loading'){
    fn();
  } else {
    document.addEventListener('DOMContentLoaded', fn);
  }
};

(function(G) {
  if (window.navigator.standalone) {
    G.ready(function() {
      /**
       * Override all link clicks, to avoid fullscreen mode opening in the
       * external browser
       */
      [].forEach.call(document.querySelectorAll("a"), function(link) {
        link.onclick = function(e) {
          e.preventDefault();

          var href = this.getAttribute("href");

          if (this.getAttribute("target") !== null) {
            window.open(href, this.getAttribute("target"))
          } else {
            window.location = href;
          }

          return false;
        };
      });

      /**
       * Adds class signifying fullscreen standalone mode for iOS devices */
      document.querySelector("body").classList.add("standalone");
    });
  }
})(Gallo);
