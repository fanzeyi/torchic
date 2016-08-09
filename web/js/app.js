{
  function escapeHTML(s) {
    return s.replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  var xhr;
  var auto_complete = new autoComplete({
    selector: '.search-input',
    minChars: 1,
    source: function(term, response) {
        try { xhr.abort(); } catch(e){}
        xhr = $.getJSON('/api/complete', { query: term }, function(data){ response(data); });
    }
  });


  var body = document.getElementsByTagName("body")[0];

  var form = document.getElementById("search-form");
  var query = $("#query-input");
  var resultList = $("#search-result-list");
  var completeList = $(".autocomplete-suggestions");

  form.addEventListener("submit", function(e) {
    $(".search-box").animate({
      height: "60px"
    }, 250, function() {
      body.classList.add("result");
      body.classList.remove("search");
    });

    completeList.css({
      display: 'none'
    });

    $.getJSON("/api/query", {
      query: query.val()
    }, function(resp) {
      resultList.empty();
      for (item of resp) {
        var entry = `<li class="search-result">
          <h2>
            <a href="${item.url}">${item.title}</a>
          </h2>
          <small class="meta">${escapeHTML(item.url)}</small>
          <div class="summary">
          </div>
        </li>`

        resultList.append(entry);
      }
    });

    e.preventDefault();
  });

}
