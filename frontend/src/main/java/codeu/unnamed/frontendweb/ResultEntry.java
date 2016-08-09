package codeu.unnamed.frontendweb;

/**
 * Created by fanzeyi on 8/9/16.
 */
public class ResultEntry {
    private String title;

    private String URL;

    private String summary;

    public ResultEntry(String title, String URL, String summary) {
        this.title = title;
        this.URL = URL;
        this.summary = summary;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }

    public String getURL() {
        return URL;
    }

    public void setURL(String URL) {
        this.URL = URL;
    }

//    public String getSummary() {
//        return summary;
//    }

    public void setSummary(String summary) {
        this.summary = summary;
    }
}
