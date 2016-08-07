package codeu.unnamed.frontendweb;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.tartarus.snowball.SnowballStemmer;
import org.tartarus.snowball.ext.englishStemmer;

import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@RestController
public class QueryController {
    @Autowired
    private AutoCompletion autoCompletion;

    @RequestMapping("/query")
    public String query(@RequestParam(value="query", required = true) String query, @RequestParam(value="offset", required = false, defaultValue = "0") String offset) {
        // TODO: process query with snowball here
        String[] terms = processQuery(query);

        // TODO: fill in here to get result from INDEX

        // TODO: wrap up the result as response

        return String.join(", ", terms);
    }

    private String[] processQuery(String query) {
        String[] terms = query.split(" ");

        SnowballStemmer stemmer = new englishStemmer();

        Stream<String> stream = Arrays.stream(query.split((" "))).map((String term) -> {
            stemmer.setCurrent(term);
            stemmer.stem();
            return stemmer.getCurrent();
        });

        return stream.toArray(String[]::new);
    }

    @RequestMapping("/complete")
    public List<String> complete(@RequestParam(value="query", required=true) String query) {
        String[] terms = query.split(" ");

        List<String> result = this.autoCompletion.complete(terms[terms.length-1]);

        result = result.stream().map((s) -> {
            terms[terms.length-1] = s;
            return String.join(" ", terms);
        }).collect(Collectors.toList());

        return result;
    }
}
