package echo;

import io.netty.channel.ChannelInitializer;
import io.netty.channel.socket.SocketChannel;

public class EchoServerInitializer extends ChannelInitializer<SocketChannel> {


    @Override
    public void initChannel(SocketChannel channel) throws Exception {
        channel.pipeline().addLast(new EchoServerHandler());
    }

}
