package codeu.unnamed.frontendweb;

import ds.tree.RadixTree;
import ds.tree.RadixTreeImpl;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.ApplicationListener;
import org.springframework.stereotype.Component;

import java.io.*;
import java.util.List;

/**
 * Created by fanzeyi on 8/5/16.
 */
@Component
public class AutoCompletion implements ApplicationListener<ApplicationReadyEvent> {
    private RadixTree<String> completionTree;

    @Override
    public void onApplicationEvent(ApplicationReadyEvent event) {
        this.completionTree = new RadixTreeImpl<>();

        ClassLoader classLoader = getClass().getClassLoader();
        File file = new File(classLoader.getResource("wordlist.txt").getFile());
        BufferedReader reader = null;

        try {
            reader = new BufferedReader(new FileReader(file));
            String line = null;
            int i = 0;

            while ((line = reader.readLine()) != null && i++ < 10000) {
                if (line.length() >= 3) {
                    this.completionTree.insert(line, line);
                }
            }
        } catch (FileNotFoundException e) {
        } catch (IOException e) {
        } finally {
            try {
                if (reader != null) {

                    reader.close();
                }
            } catch (IOException e) {
            }
        }

    }

    public List<String> complete(String key) {
        return this.completionTree.searchPrefix(key, 5);
    }
}
