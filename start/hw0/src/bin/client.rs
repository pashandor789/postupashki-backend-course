use std::io::ErrorKind;
use std::net::{TcpStream, SocketAddr, Shutdown};
use std::io::{BufWriter, Read, Result, Error, Stdout, Write, stdout};

fn main() -> Result<()> {
    const CONNECTION_ERROR: &str    = "The socket acquired by the server should be listening on localhost with port 8080.\n";
    const DECODE_ERROR: &str        = "The data should comply with UTF-8 encoding and not contain invalid bytes.\n";
    let mut out: BufWriter<Stdout> = BufWriter::new(stdout());
    
    let mut stream: TcpStream = TcpStream::connect(
        SocketAddr::from(([127, 0, 0, 1], 8080))
    ).expect(CONNECTION_ERROR);
    let mut buf: Vec<u8> = Vec::new();

    let read_amt: usize = stream.read_to_end(&mut buf)?;
    writeln!(out, "{} bytes was read from the socket", read_amt)?;
    let read_res: &str = str::from_utf8(&buf).expect(DECODE_ERROR);
    stream.shutdown(Shutdown::Read)?;
    
    if read_res != "OK\n" {
        return Err(Error::new(ErrorKind::Other,
             format!("wrong value returned. value is \'{}\'", &read_res)
            ));
    }

    Ok(())
}