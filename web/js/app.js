{
  let xhr;
  let auto_complete = new autoComplete({
    selector: '.search-input',
    minChars: 1,
    source: function(term, response) {
        try { xhr.abort(); } catch(e){}
        xhr = $.getJSON('/api/complete', { query: term }, function(data){ response(data); });
    }
  });


  let body = document.getElementsByTagName("body")[0];

  let form = document.getElementById("search-form");
  form.addEventListener("submit", (e) => {
    $(".search-box").animate({
      height: "60px"
    }, 250, function() {
      body.classList.toggle("result");
      body.classList.toggle("search");
    });
    e.preventDefault();
  });

}
