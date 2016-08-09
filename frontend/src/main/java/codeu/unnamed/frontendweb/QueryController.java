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

    @Autowired
    private DocumentDao documentDao;

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

    private static final int WINDOW_LENGTH = 50;

    private String generateSummary(long id, String query) {
        String[] queryies = query.split(" ");

        Document document = this.documentDao.findById(id);

        String[] words = document.getText().split(" ");

        int rank = 0;

        // need a circular array
        int[] mapping = new int[WINDOW_LENGTH];

        // first [length of window] words
        for( int i = 0; i < words.length && i < WINDOW_LENGTH; i++ ) {
            if (arrayContains(queryies, words[i])) {
                if (i == 0) {
                    mapping[i] = 1;
                } else {
                    mapping[i] = mapping[i-1] + 1;
                }

                rank += mapping[i];
            }
        }

        int index = 0;
        int highest = rank;

        int end = 0; // end position of circular

        for( int i = WINDOW_LENGTH; i < words.length; i++ ) {
            rank -= mapping[end % WINDOW_LENGTH];

            if (arrayContains(queryies, words[i])) {
                mapping[end] = mapping[(end-1) % WINDOW_LENGTH] + 1;
            } else {
                mapping[end] = 0;
            }

            rank += mapping[end];
            end = (end + 1) % WINDOW_LENGTH;

            if (rank > highest) {
                highest = rank;
                index = WINDOW_LENGTH + i;
            }
        }

        String[] result = new String[WINDOW_LENGTH];

        for ( int i = index; i < words.length && i < index + WINDOW_LENGTH; i++ ) {
            result[i-index] = words[i];
        }

        return String.join(" ", result);
    }

    private static <T extends Comparable> boolean arrayContains(T[] array, T target) {
        for (T element : array) {
            if (element.compareTo(target) == 0) {
                return true;
            }
        }

        return false;
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
