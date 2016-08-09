package codeu.unnamed.frontendweb;

import javax.persistence.*;
import javax.validation.constraints.NotNull;

/**
 * Created by fanzeyi on 8/3/16.
 */
@Entity
@Table(name = "urls")
public class Document {
    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private long id;

    @NotNull
    private String hash;

    @NotNull
    private String url;

    @NotNull
    private String text;

    @NotNull
    private String title;

    public Document() {
    }

    public Document(long id) {
        this.id = id;
    }

    public Document(String hash, String url) {
        this.hash = hash;
        this.url = url;
    }

    public long getId() {
        return id;
    }

    public void setId(long id) {
        this.id = id;
    }

    public String getHash() {
        return hash;
    }

    public void setHash(String hash) {
        this.hash = hash;
    }

    public String getUrl() {
        return url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }
}
