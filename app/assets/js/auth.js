var Gallo = window.Gallo || {};

(function(G, w) {
  G.ready(function() {
    if (w.location.hash !== "") {
      const regexToken = /[&#]?token=([0-9a-f]{64})/;
      const match = regexToken.exec(w.location.hash);

      if (match) {
        token = match[1];
        w.location.href = "/auth?token=" + token;
      }
    }
  });
})(Gallo, window);
