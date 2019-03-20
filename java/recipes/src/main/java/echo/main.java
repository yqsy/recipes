package echo;

public class main {
    public static void main(String[] args) {

        try {
            int port = 9999;
            //System.out.printf("listen to :%d", port);
            new EchoServer(port).run();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
