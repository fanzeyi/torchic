{
  var escapeHTML = function(s) {
    return s.replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  };

  var highlight = function(text, query) {
    var terms = query.split(/\s/);

    for (term of terms) {
      text = text.replace(new RegExp(term, 'i'), "<strong>$&</strong>");
    }

    return text;
  };

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

  var notFound = $(".not-found");
  var summary = $(".search-summary");

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

    var q = query.val();

    $.getJSON("/api/query", {
      query: q
    }, function(resp) {
      $(".search-result").remove();

      if(resp.length == 0) {
        notFound.addClass("show");
        summary.addClass("hide");
      } else{
        notFound.removeClass("show");
        summary.removeClass("hide");
      }

      for (item of resp) {
        var entry = `<li class="search-result">
          <h2><a href="${item.url}">${highlight(item.title, q)}</a></h2>
          <small class="meta">${highlight(escapeHTML(item.url), q)}</small>
          <div class="summary">${highlight(escapeHTML(item.summary), q)}</div>
        </li>`

        resultList.append(entry);
      }
    });

    e.preventDefault();
  });

}
