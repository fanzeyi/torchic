{
  let auto_complete = new autoComplete({
    selector: '.search-input',
    minChars: 1,
    source: function(term, suggest) {
      term = term.toLowerCase();
      let choices = ['hello', 'what', 'bug', 'when', 'where', 'which'];
      let matches = [];
      for (let word of choices) {
        if (word.includes(term)) matches.push(word);
      }

      suggest(matches);
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
