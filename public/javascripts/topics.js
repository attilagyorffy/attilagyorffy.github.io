(function () {
    var filter = document.querySelector(".topic-filter");
    if (!filter) return;

    var buttons = filter.querySelectorAll("[data-topic]");
    var posts = document.querySelectorAll(".post-item[data-topics]");
    var sections = document.querySelectorAll(".post-list");
    var active = null;
    var animationsCleared = false;
    var pendingTimeouts = [];
    var reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    function clearInitialAnimations() {
        if (animationsCleared) return;
        animationsCleared = true;
        document.querySelectorAll(".post-list h2, .post-list .post-item").forEach(function (el) {
            el.style.animation = "none";
            el.style.opacity = "1";
        });
    }

    function setVisibility() {
        posts.forEach(function (post) {
            var topics = post.dataset.topics.split(" ");
            var match = !active || topics.indexOf(active) !== -1;
            post.classList.toggle("hidden", !match);
        });

        sections.forEach(function (section) {
            var visible = section.querySelectorAll(".post-item:not(.hidden)");
            section.classList.toggle("hidden", !visible.length);
        });
    }

    function getVisibleElements() {
        var order = [];
        sections.forEach(function (section) {
            if (section.classList.contains("hidden")) return;
            var h2 = section.querySelector("h2");
            if (h2) order.push(h2);
            section.querySelectorAll(".post-item:not(.hidden)").forEach(function (post) {
                order.push(post);
            });
        });
        return order;
    }

    function updateURL() {
        var url = new URL(window.location);
        if (active) {
            url.searchParams.set("topic", active);
        } else {
            url.searchParams.delete("topic");
        }
        history.replaceState(null, "", url);
    }

    function clearPending() {
        pendingTimeouts.forEach(clearTimeout);
        pendingTimeouts = [];
    }

    function applyFilter(topic, animate) {
        active = active === topic ? null : topic;

        buttons.forEach(function (btn) {
            btn.setAttribute("aria-pressed", btn.dataset.topic === active ? "true" : "false");
        });
        filter.classList.toggle("has-active", !!active);

        if (!animate || reducedMotion) {
            setVisibility();
            updateURL();
            return;
        }

        clearPending();
        clearInitialAnimations();

        // Phase 1: Fade out currently visible elements
        var currentlyVisible = getVisibleElements();
        var fadeOutMs = 250;

        currentlyVisible.forEach(function (el) {
            el.style.transition = "opacity " + fadeOutMs + "ms cubic-bezier(0.4, 0, 1, 1)";
            el.style.opacity = "0";
        });

        var t1 = setTimeout(function () {
            // Set all post-list elements to opacity 0 before changing visibility
            // to prevent flash when hidden sections become visible
            // Set start state with transitions disabled to prevent
            // .post-item { transition: transform 0.2s ease } from interfering
            document.querySelectorAll(".post-list h2, .post-list .post-item").forEach(function (el) {
                el.style.transition = "none";
                el.style.opacity = "0";
                el.style.transform = "translateY(12px)";
            });

            setVisibility();

            // Force reflow so the browser registers the start state
            // before any transitions are applied
            void document.body.offsetHeight;

            // Phase 2: Staggered fade-in
            var order = getVisibleElements();
            var staggerMs = 120;
            var fadeInMs = 450;
            var easing = "cubic-bezier(0, 0, 0.2, 1)";

            order.forEach(function (el, i) {
                var t = setTimeout(function () {
                    el.style.transition = "opacity " + fadeInMs + "ms " + easing + ", transform " + fadeInMs + "ms " + easing;
                    el.style.opacity = "1";
                    el.style.transform = "translateY(0)";
                }, i * staggerMs);
                pendingTimeouts.push(t);
            });

            // Clean up inline styles after all animations finish,
            // restoring the base .post-item hover transition
            var cleanupDelay = (order.length * staggerMs) + fadeInMs + 50;
            var t2 = setTimeout(function () {
                order.forEach(function (el) {
                    el.style.transition = "";
                    el.style.transform = "";
                });
            }, cleanupDelay);
            pendingTimeouts.push(t2);
        }, fadeOutMs + 80);
        pendingTimeouts.push(t1);

        updateURL();
    }

    buttons.forEach(function (btn) {
        btn.addEventListener("click", function () {
            applyFilter(btn.dataset.topic, true);
        });
    });

    // Apply filter from URL on load (no animation, initial CSS handles reveal)
    var params = new URLSearchParams(window.location.search);
    var initial = params.get("topic");
    if (initial) {
        applyFilter(initial, false);
    }
})();
