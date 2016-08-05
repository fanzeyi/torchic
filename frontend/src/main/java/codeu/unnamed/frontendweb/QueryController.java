package codeu.unnamed.frontendweb;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.tartarus.snowball.SnowballStemmer;
import org.tartarus.snowball.ext.englishStemmer;

import java.util.Arrays;
import java.util.stream.Stream;

@RestController
public class QueryController {
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
}
