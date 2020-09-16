var Gallo = window.Gallo || {};

// Amount of time fading images in- and out takes in ms.
Gallo.FADE_DURATION = 2000;

// Amount of time each image is shown before changing again in ms.
Gallo.SHOW_DURATION = Gallo.SHOW_DURATION || 5000;

// From experimentation any document that extends outside of ~25000px in width
// will cause a crash on Safari[1]. This limit is set lower to leave some
// headroom.
Gallo.DOM_WIDTH_LIMIT = 22500;

// Minimum number of images to wait on, before considering images to be loaded.
// Since most will be loading off-screen, theres no reason to wait for all
// images to be loaded. Only enough to at least cover the entire screen, when
// the cover fades out. This value is a guesstimate for a lower bound worst
// case. E.g. if all images are portrait, there's probaly room for 3-4.
Gallo.IMAGE_LOAD_WAIT_COUNT = 5;

//------------------------------------------------------------------------------

/**
 * Simple version of [forEach()]{@link safeRemove} which will work on Safari[1].
 *
 * @see {@link https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/forEach}
 *
 * @param {Function} callback - Function to execute on each element.
 */
Gallo.fooEach = function(callback) {
  for (var i = 0; i < this.length; i++) { callback(this[i]); }
};

/**
 * Version of DOMTokenList.remove(), which correctly handles multiple
 * arguments on Safari[1].
 *
 * @see {@link https://developer.mozilla.org/en-US/docs/Web/API/DOMTokenList/remove}
 */
Gallo.safeRemove = function() {
  for(var i = 0; i < arguments.length; i++) {
    this.remove(arguments[i]);
  }
};

// Monkey patches
DOMTokenList.prototype.safeRemove = Gallo.safeRemove;
NodeList.prototype.fooEach = Gallo.fooEach;
Array.prototype.fooEach = Gallo.fooEach;

/**
 * Randomises elements of an array in-place.
 *
 * @see {@link https://en.wikipedia.org/wiki/FishereYates_shuffle}
 *
 * @param {Array} array - The array to be shuffled. Any array type with a
 *                        .length will do.
 */
Gallo.shuffle = function(array) {
  console.assert(!!array, 'shuffle: array is not truthy');

  var currentIndex = array.length, temporaryValue, randomIndex;

  while (0 !== currentIndex) {
    randomIndex = Math.floor(Math.random() * currentIndex);
    currentIndex -= 1;

    temporaryValue = array[currentIndex];
    array[currentIndex] = array[randomIndex];
    array[randomIndex] = temporaryValue;
  }

  return array;
};

/**
 * Finds the correct event name for transition end depending on the browser.
 *
 * @see {@link https://jonsuh.com/blog/detect-the-end-of-css-animations-and-transitions-with-javascript/}
 * @returns {string}
 */
Gallo.whichTransitionEvent = (function() {
  var value;

  return function() {
    if (!value) {
      var el = document.createElement("fakeelement");

      var transitions = {
        "transition"      : "transitionend",
        "OTransition"     : "oTransitionEnd",
        "MozTransition"   : "transitionend",
        "WebkitTransition": "webkitTransitionEnd"
      }

      for (var t in transitions){
        if (el.style[t] !== undefined){
          value = transitions[t];
          break;
        }
      }
    }

    return value;
  };
})();

/**
 * Finds the correct, possibly prefixed, version of the transform style
 * property.
 *
 * @returns {string}
 */
Gallo.whichTransform = (function() {
  var value;

  return function() {
    if (!value) {
      var el = document.createElement("fakeelement");

      var transforms = {
        'transform'      : 'transform',
        'WebkitTransform': '-webkit-transform',
        'msTransform'    : '-ms-transform',
        'MozTransform'   : '-moz-transform',
        'OTransform'     : '-o-transform'
      };

      for (var t in transforms){
        if (el.style[t] !== undefined){
          value = transforms[t];
          break;
        }
      }
    }

    return value;
  };
})();

/**
 * Generates the appropriate value for the srset image attribute, based on
 * image and viewport dimensions.
 *
 * @param {object} image - Image data object with previews property.
 * @returns {string}
 */
Gallo.srcSet = function(image) {
  var tmp = [];

  image.previews.fooEach(function(preview) {
    tmp.push(preview.url.concat(' ', preview.width.toString(), 'w'));
  });

  return tmp.join(', ');
};

/**
 * Generates the appropriate value for the sizes image attribute, based on
 * image and viewport dimensions.
 *
 * @param {object} image - Image data object with previews property.
 * @returns {string}
 */
Gallo.sizes = function(image) {
  var imageRatio, viewportRatio;
  var lastPreview = image.previews[image.previews.length - 1];
  var body = document.querySelector('body');

  // If the image is a portrait, then it most likely doesn't fit the aspect
  // ratio of the viewport. This leaves empty space on either side of the
  // image. The calculation determines the fraction of the viewport width
  // taken up by the image.
  if (lastPreview.width < lastPreview.height) {
    imageRatio = lastPreview.width / lastPreview.height;
    viewportRatio = body.offsetWidth / body.offsetHeight;

    return Math.round((imageRatio / viewportRatio) * 100) + "vw";
  }

  return "100vw";
};

/**
 * Manages showing the images defined on the Gallo.images data property. It
 * takes care of fading from cover to the image view, as well as the animation
 * and change between individual images.
 *
 * @param {Object} G Gallo root object
 * @param {Document} d Global document
 * @param {Window} w Global window
 * @param {console} c Global console
 */
Gallo.present = function(G, d, w, c) {
  var imagesEl = d.querySelector('.images') ||
    c.assert(!!imagesEl, 'Images element not found!');
  var coverEl = d.querySelector('.cover') ||
    c.assert(!!coverEl, 'Cover container not found!');
  c.assert(!!G.IMAGES, 'Gallo.images data property not set!');

  //----------------------------------------------------------------------------

  var state;
  var transform = G.whichTransform();
  var transitionEvent = G.whichTransitionEvent();
  var presentationWidth = d.querySelector('body').offsetWidth;

  // Shuffle previews so the order is different on each load
  var shuffledImages = G.shuffle(G.IMAGES);
  var preview, image, imageElements = [], imagesLoadedCounter = 0, div;
  var totalWidth = 0;

  for(var i = 0; i < shuffledImages.length; i++) {
    image = shuffledImages[i];

    // Since the image is going to take up the entire viewport height, we can
    // calculate the width it's going to have from the aspect ratio.
    var lastPreview = image.previews[image.previews.length - 1];
    var imageAspectRatio = lastPreview.width / lastPreview.height;
    var imageRenderedWidth = d.documentElement.clientHeight *
      (lastPreview.width / lastPreview.height);

    // See comment for DOM_WIDTH_LIMIT
    if (totalWidth + imageRenderedWidth > G.DOM_WIDTH_LIMIT) {
      break;
    }

    imageEl = new Image();
    imageEl.setAttribute('srcset', G.srcSet(image));
    imageEl.setAttribute('sizes', G.sizes(image));
    imageEl.classList.add('image');

    /**
     * This should really have been a 'height: 100%;' on the css side of things,
     * but due to some weird behaviour on Safari[1] when doing transforms, which
     * results in an ever increasing zoom level, this manual setting of the
     * height directly on the element, is the effective workaround.
     */
    imageEl.style.cssText = 'height: ' + d.documentElement.clientHeight + 'px;';

    imageEl.addEventListener('load', function() {
      if(++imagesLoadedCounter >= Math.min(
        G.IMAGE_LOAD_WAIT_COUNT,
        imageElements.length
      )) {
        state.imagesLoad();
      }
    });

    imageEl.addEventListener('onerror', console.error);

    imageElements.push(imageEl);
    imagesEl.appendChild(imageEl);

    totalWidth += imageRenderedWidth;
  }

  state = new StateMachine({
    transitions: [
      { name: 'coverTimeout', from: 'none', to: 'coverTimedOut' },
      { name: 'imagesLoad', from: 'none', to: 'imagesLoaded' },
      // This extra transition is here, because the total number of images to
      // load is typically greater than the minimum limit required to trigger
      // the transition initally. Hence any subsequent loads should lead to the
      // same state.
      { name: 'imagesLoad', from: 'imagesLoaded', to: 'imagesLoaded' },
      { name: 'coverTimeout', from: 'imagesLoaded', to: 'fadingOutCover' },
      { name: 'imagesLoad', from: 'coverTimedOut', to: 'fadingOutCover' },
      { name: 'doneFadingOutCover', from: 'fadingOutCover', to: 'fadingInImages' },
      { name: 'doneFadingInImages', from: 'fadingInLayerZero', to: 'imagesShowing' },
    ],
    methods: {
      onFadingOutCover: function() {
        var that = this;

        var onDoneFadingOut = function() {
          that.doneFadingOutCover();

          that.coverEl.removeEventListener(that.transitionEvent, onDoneFadingOut);
        };

        this.coverEl.addEventListener(this.transitionEvent, onDoneFadingOut);

        this.coverEl.classList.add('transparent');
      },
      onFadingInImages: function() {
        var that = this;
        var currentImageIndex = 0;

        var moveToNextImage = function() {
          var currentImage = that.imageElements[currentImageIndex];
          var x;

          if (currentImageIndex === 0) {
            // Align left edge of first image with left edge of viewport
            x = 0;
          } else if (currentImageIndex === that.imageElements.length - 1) {
            // Align right edge of last image with right edge of viewport
            x = -currentImage.offsetLeft + (presentationWidth - currentImage.width);
          } else {
            // Align horizontal center of other images with center viewport
            x = -currentImage.offsetLeft + (presentationWidth - currentImage.width) / 2;
          }

          that.imagesEl.style.cssText = transform + ': translate3d(' + x + 'px, 0, 0)';

          currentImage.classList.add('focus');
          setTimeout(function() {
            currentImage.classList.remove('focus');
          }, G.SHOW_DURATION);

          currentImageIndex = (currentImageIndex + 1) % that.imageElements.length;
        }

        this.imagesEl.classList.safeRemove('transparent', 'hidden');

        setInterval(moveToNextImage, G.SHOW_DURATION);
        moveToNextImage();
      },
    },
    data: {
      coverEl: coverEl,
      imagesEl: imagesEl,
      imageElements: imageElements,
      transitionEvent: transitionEvent
    }
  });

  picturefill({ reevaluate: true, elements: imageElements });

  // Let the cover stay for 10 seconds before beginning to cycle imageElements
  setTimeout(function() { state.coverTimeout(); }, 10000);
};

Gallo.ready(function() {
  Gallo.present(Gallo, document, window, console);

  if (Gallo.REFRESH) {
    setTimeout(function() { location.reload(); }, Gallo.REFRESH);
  }
});

/**
 * Reference
 *
 * 1. Specifically Safari on iOS 5.1.1, which is the last version available on
 * the 1st generation iPad.
 */
