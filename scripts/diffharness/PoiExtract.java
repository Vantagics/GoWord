import java.io.FileInputStream;
import org.apache.poi.xwpf.extractor.XWPFWordExtractor;
import org.apache.poi.xwpf.usermodel.XWPFDocument;

/** Prints the text Apache POI XWPF extracts from a .docx file. */
public class PoiExtract {
    public static void main(String[] args) throws Exception {
        if (args.length != 1) {
            System.err.println("usage: PoiExtract <file.docx>");
            System.exit(2);
        }
        try (FileInputStream in = new FileInputStream(args[0]);
             XWPFDocument doc = new XWPFDocument(in);
             XWPFWordExtractor ex = new XWPFWordExtractor(doc)) {
            String text = ex.getText();
            if (text != null) {
                System.out.print(text);
            }
        }
    }
}
