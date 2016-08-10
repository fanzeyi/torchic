package codeu.unnamed.frontendweb;

/**
 * Created by fanzeyi on 8/9/16.
 */
public class ResultEntry {
    private Document document;
    
    private String[] queries;

    private String summary;

    private static final int WINDOW_LENGTH = 50;

    public ResultEntry(Document document, String[] query) {
        this.document = document;
        this.queries = query;

        this.summary = generateSummary();
    }

    public String getTitle() {
        return this.document.getTitle();
    }

    public String getURL() {
        return this.document.getUrl();
    }

    public String getSummary() {
        return summary;
    }

    private static int circularIndex(int i) {
        if(i < 0) {
            return (i + WINDOW_LENGTH);
        }
        return i % WINDOW_LENGTH;
    }

    public String generateSummary() {
        String[] words = document.getText().split(" ");

        int rank = 0;

        // need a circular array
        int[] mapping = new int[WINDOW_LENGTH];

        // first [length of window] words
        for( int i = 0; i < words.length && i < WINDOW_LENGTH; i++ ) {
            if (arrayContains(queries, words[i])) {
                if (i == 0) {
                    mapping[i] = 1;
                } else {
                    mapping[i] = mapping[i-1] + 1;
                }

                rank += mapping[i];
            }
        }

        if( words.length < WINDOW_LENGTH ) {
            return String.join(" ", words);
        }

        int index = 0;
        int highest = rank;

        int end = 0; // end position of circular

        for( int i = WINDOW_LENGTH; i < words.length; i++ ) {
            rank -= mapping[end];

            if (arrayContains(queries, words[i])) {
                mapping[end] = mapping[circularIndex(end-1)] + 1;
            } else {
                mapping[end] = 0;
            }

            rank += mapping[end];
            end = circularIndex(end + 1);

            if (rank > highest) {
                highest = rank;
                index = i;
            }
        }

        if (index > 25) {
            index -= 25;
        } else {
            index = 0;
        }

        String[] result = new String[WINDOW_LENGTH];

        for ( int i = index; i < words.length && i < index + WINDOW_LENGTH; i++ ) {
            result[i-index] = words[i];
        }

        return String.join(" ", result);
    }

    private static boolean arrayContains(String[] array, String target) {
        for (String element : array) {
            if (target.equalsIgnoreCase(element)) {
                return true;
            }
        }

        return false;
    }
}
