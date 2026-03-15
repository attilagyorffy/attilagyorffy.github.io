(function () {
    var lightbox = document.getElementById("lightbox");
    var imgA = document.getElementById("lightbox-img-a");
    var imgB = document.getElementById("lightbox-img-b");
    var counter = document.getElementById("lightbox-counter");
    var caption = document.getElementById("lightbox-caption");
    var links = Array.from(
        document.querySelectorAll(".photo-link"),
    );
    var captions = Array.from(
        document.querySelectorAll(".photo-caption"),
    );
    var currentIndex = 0;
    var activeImg = imgA;
    var navigating = false;
    var preloadCache = {};

    function preload(index) {
        for (var offset = 1; offset <= 2; offset++) {
            var next = (index + offset) % links.length;
            var prev = (index - offset + links.length) % links.length;
            if (!preloadCache[next]) {
                preloadCache[next] = new Image();
                preloadCache[next].src = links[next].href;
            }
            if (!preloadCache[prev]) {
                preloadCache[prev] = new Image();
                preloadCache[prev].src = links[prev].href;
            }
        }
    }

    function open(index) {
        currentIndex = index;
        activeImg = imgA;
        imgA.style.transition = "none";
        imgA.style.opacity = "1";
        imgA.style.transform = "translateX(0)";
        imgA.src = links[index].href;
        imgB.style.opacity = "0";
        imgB.src = "";
        counter.textContent =
            index + 1 + " / " + links.length;
        caption.textContent = captions[index]
            ? captions[index].textContent
            : "";
        caption.style.opacity = "1";
        lightbox.setAttribute("aria-hidden", "false");
        document.body.style.overflow = "hidden";
        history.replaceState(null, "", "#photo-" + (index + 1));
        requestAnimationFrame(function () {
            imgA.style.transition = "";
        });
        preload(index);
    }

    function navigate(index) {
        if (navigating) return;
        navigating = true;
        var outgoing = activeImg;
        var incoming = activeImg === imgA ? imgB : imgA;
        activeImg = incoming;
        currentIndex = index;

        // prepare incoming: off to the right, invisible
        incoming.style.transition = "none";
        incoming.style.opacity = "0";
        incoming.style.transform = "translateX(16px)";
        incoming.src = links[index].href;

        function startTransition() {
            incoming.offsetHeight; // force layout
            incoming.style.transition = "";
            // slide old image out left + fade out
            outgoing.style.opacity = "0";
            outgoing.style.transform = "translateX(-16px)";
            caption.style.opacity = "0";
            // slide new image in
            incoming.style.opacity = "1";
            incoming.style.transform = "translateX(0)";
            counter.textContent =
                index + 1 + " / " + links.length;
            history.replaceState(null, "", "#photo-" + (index + 1));
            setTimeout(function () {
                caption.textContent = captions[index]
                    ? captions[index].textContent
                    : "";
                caption.style.opacity = "1";
                outgoing.src = "";
                navigating = false;
                preload(index);
            }, 250);
        }

        if (incoming.complete && incoming.naturalWidth > 0) {
            startTransition();
        } else {
            incoming.onload = function () {
                incoming.onload = null;
                startTransition();
            };
        }
    }

    function close() {
        lightbox.setAttribute("aria-hidden", "true");
        imgA.src = "";
        imgB.src = "";
        document.body.style.overflow = "";
        navigating = false;
        history.replaceState(null, "", location.pathname);
    }

    function prev() {
        navigate(
            (currentIndex - 1 + links.length) % links.length,
        );
    }

    function next() {
        navigate((currentIndex + 1) % links.length);
    }

    links.forEach(function (link, i) {
        link.addEventListener("click", function (e) {
            e.preventDefault();
            open(i);
        });
    });

    lightbox
        .querySelector(".lightbox-close")
        .addEventListener("click", close);
    lightbox
        .querySelector(".lightbox-prev")
        .addEventListener("click", prev);
    lightbox
        .querySelector(".lightbox-next")
        .addEventListener("click", next);

    lightbox.addEventListener("click", function (e) {
        if (
            e.target === lightbox ||
            e.target.classList.contains("lightbox-content")
        ) {
            close();
        }
    });

    document.addEventListener("keydown", function (e) {
        if (lightbox.getAttribute("aria-hidden") === "true")
            return;
        if (e.key === "Escape") close();
        if (e.key === "ArrowLeft") prev();
        if (e.key === "ArrowRight") next();
    });

    var touchStartX = 0;
    lightbox.addEventListener(
        "touchstart",
        function (e) {
            touchStartX = e.changedTouches[0].screenX;
        },
        { passive: true },
    );
    lightbox.addEventListener(
        "touchend",
        function (e) {
            var diff =
                e.changedTouches[0].screenX - touchStartX;
            if (Math.abs(diff) > 50) {
                diff > 0 ? prev() : next();
            }
        },
        { passive: true },
    );

    // Open lightbox from permalink hash
    var match = location.hash.match(/^#photo-(\d+)$/);
    if (match) {
        var photoIndex = parseInt(match[1], 10) - 1;
        if (photoIndex >= 0 && photoIndex < links.length) {
            open(photoIndex);
        }
    }
})();
