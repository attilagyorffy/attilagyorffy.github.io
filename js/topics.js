(function () {
    var filter = document.querySelector(".topic-filter");
    if (!filter) return;

    var buttons = filter.querySelectorAll("[data-topic]");
    var posts = document.querySelectorAll(".post-item[data-topics]");
    var sections = document.querySelectorAll(".post-list");
    var active = null;

    function applyFilter(topic) {
        active = active === topic ? null : topic;

        buttons.forEach(function (btn) {
            btn.setAttribute("aria-pressed", btn.dataset.topic === active ? "true" : "false");
        });

        posts.forEach(function (post) {
            var topics = post.dataset.topics.split(" ");
            var match = !active || topics.indexOf(active) !== -1;
            post.classList.toggle("hidden", !match);
        });

        // Hide year sections that have no visible posts
        sections.forEach(function (section) {
            var visible = section.querySelectorAll(".post-item:not(.hidden)");
            section.classList.toggle("hidden", !visible.length);
        });

        // Update URL without reload
        var url = new URL(window.location);
        if (active) {
            url.searchParams.set("topic", active);
        } else {
            url.searchParams.delete("topic");
        }
        history.replaceState(null, "", url);
    }

    buttons.forEach(function (btn) {
        btn.addEventListener("click", function () {
            applyFilter(btn.dataset.topic);
        });
    });

    // Apply filter from URL on load
    var params = new URLSearchParams(window.location.search);
    var initial = params.get("topic");
    if (initial) {
        applyFilter(initial);
    }
})();
