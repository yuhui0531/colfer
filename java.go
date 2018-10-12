package colfer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const javaKeywords = "abstract assert boolean break byte case catch char class const continue default do double else enum extends final finally float for goto if implements import instanceof int interface long native new package private protected public return short static strictfp super switch synchronized this throw throws transient try void volatile while"

// IsJavaKeyword returs whether s is a reserved word in Java code.
func IsJavaKeyword(s string) bool {
	for _, k := range strings.Fields(javaKeywords) {
		if k == s {
			return true
		}
	}
	return false
}

// GenerateJava writes the code into the respective ".java" files.
func GenerateJava(basedir string, packages Packages) error {
	packageTemplate := template.New("java-package")
	template.Must(packageTemplate.Parse(javaPackage))
	codeTemplate := template.New("java-code")
	template.Must(codeTemplate.Parse(javaCode))

	for _, p := range packages {
		var buf bytes.Buffer
		for i, seg := range strings.Split(p.Name, "/") {
			if i != 0 {
				buf.WriteByte('.')
			}
			buf.WriteString(strings.ToLower(seg))
			if IsJavaKeyword(seg) {
				buf.WriteByte('_')
			}
		}
		p.NameNative = buf.String()

		buf.Reset()
		for i, seg := range strings.Split(p.SuperClass, "/") {
			if i != 0 {
				buf.WriteByte('.')
			}
			buf.WriteString(seg)
			if IsJavaKeyword(seg) {
				buf.WriteByte('_')
			}
		}
		p.SuperClassNative = buf.String()
	}

	for _, p := range packages {
		pkgdir := filepath.Join(basedir, strings.Replace(p.NameNative, ".", string([]rune{filepath.Separator}), -1))
		if err := os.MkdirAll(pkgdir, os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		if doc := p.DocText(" * "); doc != "" {
			f, err := os.Create(filepath.Join(pkgdir, "package-info.java"))
			if err != nil {
				return err
			}
			defer f.Close()

			if err := packageTemplate.Execute(f, p); err != nil {
				return err
			}
		}

		for _, s := range p.Structs {
			for _, f := range s.Fields {
				switch f.Type {
				default:
					if f.TypeRef == nil {
						f.TypeNative = f.Type
					} else {
						f.TypeNative = f.TypeRef.NameTitle()
						if f.TypeRef.Pkg != p {
							f.TypeNative = f.TypeRef.Pkg.NameNative + "." + f.TypeNative
						}
					}
				case "bool":
					f.TypeNative = "boolean"
				case "uint8":
					f.TypeNative = "byte"
				case "uint16":
					f.TypeNative = "short"
				case "uint32", "int32":
					f.TypeNative = "int"
				case "uint64", "int64":
					f.TypeNative = "long"
				case "float32":
					f.TypeNative = "float"
				case "float64":
					f.TypeNative = "double"
				case "timestamp":
					f.TypeNative = "java.time.Instant"
				case "text":
					f.TypeNative = "String"
				case "binary":
					f.TypeNative = "byte[]"
				}

				f.NameNative = f.Name
				if IsJavaKeyword(f.NameNative) {
					f.NameNative += "_"
				}
			}

			f, err := os.Create(filepath.Join(pkgdir, s.NameTitle()+".java"))
			if err != nil {
				return err
			}
			defer f.Close()

			if err := codeTemplate.Execute(f, s); err != nil {
				return err
			}
		}
	}
	return nil
}

const javaPackage = `// Code generated by colf(1); DO NOT EDIT.
// The compiler used schema file {{.SchemaFileList}}.

/**
{{.DocText " * "}}
 */
package {{.NameNative}};
`

const javaCode = `package {{.Pkg.NameNative}};


// Code generated by colf(1); DO NOT EDIT.
// The compiler used schema file {{.Pkg.SchemaFileList}}.


import static java.lang.String.format;
import java.io.IOException;
import java.io.InputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.io.ObjectStreamException;
import java.io.OutputStream;
import java.io.Serializable;
{{- if .HasText}}
import java.nio.charset.StandardCharsets;
{{- end}}
import java.util.InputMismatchException;
import java.nio.BufferOverflowException;
import java.nio.BufferUnderflowException;


/**
 * Data bean with built-in serialization support.
{{.DocText " * "}}
 * @author generated by colf(1)
 * @see <a href="https://github.com/pascaldekloe/colfer">Colfer's home</a>
 */
{{$class := .NameTitle}}public class {{$class}}{{if .Pkg.SuperClassNative}} extends {{.Pkg.SuperClassNative}}{{end}} implements Serializable {

	/** The upper limit for serial byte sizes. */
	public static int colferSizeMax = {{.Pkg.SizeMax}};
{{if .HasList}}
	/** The upper limit for the number of elements in a list. */
	public static int colferListMax = {{.Pkg.ListMax}};
{{end}}
{{- range .Fields}}
{{if .Docs}}
	/**
{{.DocText "\t * "}}
	 */
{{- end}}
	public {{.TypeNative}}{{if .TypeList}}[]{{end}} {{.NameNative}};{{end}}


	/** Default constructor */
	public {{$class}}() {
		init();
	}
{{if .HasBinary}}
	private static final byte[] _zeroBytes = new byte[0];
{{- end}}
{{- if .HasBinaryList}}
	private static final byte[][] _zeroBinaries = new byte[0][];
{{- end}}
{{- range .Fields}}
{{- if .TypeList}}
 {{- if ne .Type "binary"}}
	private static final {{.TypeNative}}[] _zero{{.NameTitle}} = new {{.TypeNative}}[0];
 {{- end}}
{{- end}}
{{- end}}

	/** Colfer zero values. */
	private void init() {
{{- range .Fields}}
{{- if eq .Type "binary"}}
  {{- if .TypeList}}
		{{.NameNative}} = _zeroBinaries;
  {{- else}}
		{{.NameNative}} = _zeroBytes;
{{- end}}
{{- else if .TypeList}}
		{{.NameNative}} = _zero{{.NameTitle}};
{{- else if eq .Type "text"}}
		{{.NameNative}} = "";
{{- end}}
{{- end}}
	}

	/**
	 * {@link #reset(InputStream) Reusable} deserialization of Colfer streams.
	 */
	public static class Unmarshaller {

		/** The data source. */
		protected InputStream in;

		/** The read buffer. */
		public byte[] buf;

		/** The {@link #buf buffer}'s data start index, inclusive. */
		protected int offset;

		/** The {@link #buf buffer}'s data end index, exclusive. */
		protected int i;


		/**
		 * @param in the data source or {@code null}.
		 * @param buf the initial buffer or {@code null}.
		 */
		public Unmarshaller(InputStream in, byte[] buf) {
			// TODO: better size estimation
			if (buf == null || buf.length == 0)
				buf = new byte[Math.min({{$class}}.colferSizeMax, 2048)];
			this.buf = buf;
			reset(in);
		}

		/**
		 * Reuses the marshaller.
		 * @param in the data source or {@code null}.
		 * @throws IllegalStateException on pending data.
		 */
		public void reset(InputStream in) {
			if (this.i != this.offset) throw new IllegalStateException("colfer: pending data");
			this.in = in;
			this.offset = 0;
			this.i = 0;
		}

		/**
		 * Deserializes the following object.
		 * @return the result or {@code null} when EOF.
		 * @throws IOException from the input stream.
		 * @throws SecurityException on an upper limit breach defined by{{if .HasList}} either{{end}} {@link #colferSizeMax}{{if .HasList}} or {@link #colferListMax}{{end}}.
		 * @throws InputMismatchException when the data does not match this object's schema.
		 */
		public {{$class}} next() throws IOException {
			if (in == null) return null;

			while (true) {
				if (this.i > this.offset) {
					try {
						{{$class}} o = new {{$class}}();
						this.offset = o.unmarshal(this.buf, this.offset, this.i);
						return o;
					} catch (BufferUnderflowException e) {
					}
				}
				// not enough data

				if (this.i <= this.offset) {
					this.offset = 0;
					this.i = 0;
				} else if (i == buf.length) {
					byte[] src = this.buf;
					// TODO: better size estimation
					if (offset == 0) this.buf = new byte[Math.min({{$class}}.colferSizeMax, this.buf.length * 4)];
					System.arraycopy(src, this.offset, this.buf, 0, this.i - this.offset);
					this.i -= this.offset;
					this.offset = 0;
				}
				assert this.i < this.buf.length;

				int n = in.read(buf, i, buf.length - i);
				if (n < 0) {
					if (this.i > this.offset)
						throw new InputMismatchException("colfer: pending data with EOF");
					return null;
				}
				assert n > 0;
				i += n;
			}
		}

	}


	/**
	 * Serializes the object.
{{- range .Fields}}{{if .TypeList}}{{if eq .Type "float32" "float64"}}{{else}}
	 * All {@code null} elements in {@link #{{.NameNative}}} will be replaced with {{if eq .Type "text"}}{@code ""}{{else if eq .Type "binary"}}an empty byte array{{else}}a {@code new} value{{end}}.
{{- end}}{{end}}{{end}}
	 * @param out the data destination.
	 * @param buf the initial buffer or {@code null}.
	 * @return the final buffer. When the serial fits into {@code buf} then the return is {@code buf}.
	 *  Otherwise the return is a new buffer, large enough to hold the whole serial.
	 * @throws IOException from {@code out}.
	 * @throws IllegalStateException on an upper limit breach defined by{{if .HasList}} either{{end}} {@link #colferSizeMax}{{if .HasList}} or {@link #colferListMax}{{end}}.
	 */
	public byte[] marshal(OutputStream out, byte[] buf) throws IOException {
		// TODO: better size estimation
		if (buf == null || buf.length == 0)
			buf = new byte[Math.min({{$class}}.colferSizeMax, 2048)];

		while (true) {
			int i;
			try {
				i = marshal(buf, 0);
			} catch (BufferOverflowException e) {
				buf = new byte[Math.min({{$class}}.colferSizeMax, buf.length * 4)];
				continue;
			}

			out.write(buf, 0, i);
			return buf;
		}
	}

	/**
	 * Serializes the object.
{{- range .Fields}}{{if .TypeList}}{{if eq .Type "float32" "float64"}}{{else}}
	 * All {@code null} elements in {@link #{{.NameNative}}} will be replaced with {{if eq .Type "text"}}{@code ""}{{else if eq .Type "binary"}}an empty byte array{{else}}a {@code new} value{{end}}.
{{- end}}{{end}}{{end}}
	 * @param buf the data destination.
	 * @param offset the initial index for {@code buf}, inclusive.
	 * @return the final index for {@code buf}, exclusive.
	 * @throws BufferOverflowException when {@code buf} is too small.
	 * @throws IllegalStateException on an upper limit breach defined by{{if .HasList}} either{{end}} {@link #colferSizeMax}{{if .HasList}} or {@link #colferListMax}{{end}}.
	 */
	public int marshal(byte[] buf, int offset) {
		int i = offset;

		try {
{{- range .Fields}}{{if eq .Type "bool"}}
			if (this.{{.NameNative}}) {
				buf[i++] = (byte) {{.Index}};
			}
{{else if eq .Type "uint8"}}
			if (this.{{.NameNative}} != 0) {
				buf[i++] = (byte) {{.Index}};
				buf[i++] = this.{{.NameNative}};
			}
{{else if eq .Type "uint16"}}
			if (this.{{.NameNative}} != 0) {
				short x = this.{{.NameNative}};
				if ((x & (short)0xff00) != 0) {
					buf[i++] = (byte) {{.Index}};
					buf[i++] = (byte) (x >>> 8);
				} else {
					buf[i++] = (byte) ({{.Index}} | 0x80);
				}
				buf[i++] = (byte) x;
			}
{{else if eq .Type "uint32"}}
			if (this.{{.NameNative}} != 0) {
				int x = this.{{.NameNative}};
				if ((x & ~((1 << 21) - 1)) != 0) {
					buf[i++] = (byte) ({{.Index}} | 0x80);
					buf[i++] = (byte) (x >>> 24);
					buf[i++] = (byte) (x >>> 16);
					buf[i++] = (byte) (x >>> 8);
				} else {
					buf[i++] = (byte) {{.Index}};
					while (x > 0x7f) {
						buf[i++] = (byte) (x | 0x80);
						x >>>= 7;
					}
				}
				buf[i++] = (byte) x;
			}
{{else if eq .Type "uint64"}}
			if (this.{{.NameNative}} != 0) {
				long x = this.{{.NameNative}};
				if ((x & ~((1L << 49) - 1)) != 0) {
					buf[i++] = (byte) ({{.Index}} | 0x80);
					buf[i++] = (byte) (x >>> 56);
					buf[i++] = (byte) (x >>> 48);
					buf[i++] = (byte) (x >>> 40);
					buf[i++] = (byte) (x >>> 32);
					buf[i++] = (byte) (x >>> 24);
					buf[i++] = (byte) (x >>> 16);
					buf[i++] = (byte) (x >>> 8);
					buf[i++] = (byte) (x);
				} else {
					buf[i++] = (byte) {{.Index}};
					while (x > 0x7fL) {
						buf[i++] = (byte) (x | 0x80);
						x >>>= 7;
					}
					buf[i++] = (byte) x;
				}
			}
{{else if eq .Type "int32"}}
			if (this.{{.NameNative}} != 0) {
				int x = this.{{.NameNative}};
				if (x < 0) {
					x = -x;
					buf[i++] = (byte) ({{.Index}} | 0x80);
				} else
					buf[i++] = (byte) {{.Index}};
				while ((x & ~0x7f) != 0) {
					buf[i++] = (byte) (x | 0x80);
					x >>>= 7;
				}
				buf[i++] = (byte) x;
			}
{{else if eq .Type "int64"}}
			if (this.{{.NameNative}} != 0) {
				long x = this.{{.NameNative}};
				if (x < 0) {
					x = -x;
					buf[i++] = (byte) ({{.Index}} | 0x80);
				} else
					buf[i++] = (byte) {{.Index}};
				for (int n = 0; n < 8 && (x & ~0x7fL) != 0; n++) {
					buf[i++] = (byte) (x | 0x80);
					x >>>= 7;
				}
				buf[i++] = (byte) x;
			}
{{else if eq .Type "float32"}}
 {{- if .TypeList}}
			if (this.{{.NameNative}}.length != 0) {
				buf[i++] = (byte) {{.Index}};
				float[] a = this.{{.NameNative}};

				int l = a.length;
				if (l > {{$class}}.colferListMax)
					throw new IllegalStateException(format("colfer: {{.String}} length %d exceeds %d elements", l, {{$class}}.colferListMax));
				while (l > 0x7f) {
					buf[i++] = (byte) (l | 0x80);
					l >>>= 7;
				}
				buf[i++] = (byte) l;

				for (float f : a) {
					int x = Float.floatToRawIntBits(f);
					buf[i++] = (byte) (x >>> 24);
					buf[i++] = (byte) (x >>> 16);
					buf[i++] = (byte) (x >>> 8);
					buf[i++] = (byte) (x);
				}
			}
 {{- else}}
			if (this.{{.NameNative}} != 0.0f) {
				buf[i++] = (byte) {{.Index}};
				int x = Float.floatToRawIntBits(this.{{.NameNative}});
				buf[i++] = (byte) (x >>> 24);
				buf[i++] = (byte) (x >>> 16);
				buf[i++] = (byte) (x >>> 8);
				buf[i++] = (byte) (x);
			}
 {{- end}}
{{else if eq .Type "float64"}}
 {{- if .TypeList}}
			if (this.{{.NameNative}}.length != 0) {
				buf[i++] = (byte) {{.Index}};
				double[] a = this.{{.NameNative}};

				int l = a.length;
				if (l > {{$class}}.colferListMax)
					throw new IllegalStateException(format("colfer: {{.String}} length %d exceeds %d elements", l, {{$class}}.colferListMax));
				while (l > 0x7f) {
					buf[i++] = (byte) (l | 0x80);
					l >>>= 7;
				}
				buf[i++] = (byte) l;

				for (double f : a) {
					long x = Double.doubleToRawLongBits(f);
					buf[i++] = (byte) (x >>> 56);
					buf[i++] = (byte) (x >>> 48);
					buf[i++] = (byte) (x >>> 40);
					buf[i++] = (byte) (x >>> 32);
					buf[i++] = (byte) (x >>> 24);
					buf[i++] = (byte) (x >>> 16);
					buf[i++] = (byte) (x >>> 8);
					buf[i++] = (byte) (x);
				}
			}
 {{- else}}
			if (this.{{.NameNative}} != 0.0) {
				buf[i++] = (byte) {{.Index}};
				long x = Double.doubleToRawLongBits(this.{{.NameNative}});
				buf[i++] = (byte) (x >>> 56);
				buf[i++] = (byte) (x >>> 48);
				buf[i++] = (byte) (x >>> 40);
				buf[i++] = (byte) (x >>> 32);
				buf[i++] = (byte) (x >>> 24);
				buf[i++] = (byte) (x >>> 16);
				buf[i++] = (byte) (x >>> 8);
				buf[i++] = (byte) (x);
			}
 {{- end}}
{{else if eq .Type "timestamp"}}
			if (this.{{.NameNative}} != null) {
				long s = this.{{.NameNative}}.getEpochSecond();
				int ns = this.{{.NameNative}}.getNano();
				if (s != 0 || ns != 0) {
					if (s >= 0 && s < (1L << 32)) {
						buf[i++] = (byte) {{.Index}};
						buf[i++] = (byte) (s >>> 24);
						buf[i++] = (byte) (s >>> 16);
						buf[i++] = (byte) (s >>> 8);
						buf[i++] = (byte) (s);
						buf[i++] = (byte) (ns >>> 24);
						buf[i++] = (byte) (ns >>> 16);
						buf[i++] = (byte) (ns >>> 8);
						buf[i++] = (byte) (ns);
					} else {
						buf[i++] = (byte) ({{.Index}} | 0x80);
						buf[i++] = (byte) (s >>> 56);
						buf[i++] = (byte) (s >>> 48);
						buf[i++] = (byte) (s >>> 40);
						buf[i++] = (byte) (s >>> 32);
						buf[i++] = (byte) (s >>> 24);
						buf[i++] = (byte) (s >>> 16);
						buf[i++] = (byte) (s >>> 8);
						buf[i++] = (byte) (s);
						buf[i++] = (byte) (ns >>> 24);
						buf[i++] = (byte) (ns >>> 16);
						buf[i++] = (byte) (ns >>> 8);
						buf[i++] = (byte) (ns);
					}
				}
			}
{{else if eq .Type "text"}}
 {{- if .TypeList}}
			if (this.{{.NameNative}}.length != 0) {
				buf[i++] = (byte) {{.Index}};
				String[] a = this.{{.NameNative}};

				int x = a.length;
				if (x > {{$class}}.colferListMax)
					throw new IllegalStateException(format("colfer: {{.String}} length %d exceeds %d elements", x, {{$class}}.colferListMax));
				while (x > 0x7f) {
					buf[i++] = (byte) (x | 0x80);
					x >>>= 7;
				}
				buf[i++] = (byte) x;

				for (int ai = 0; ai < a.length; ai++) {
					String s = a[ai];
					if (s == null) {
						s = "";
						a[ai] = s;
					}

					int start = ++i;

					for (int sIndex = 0, sLength = s.length(); sIndex < sLength; sIndex++) {
						char c = s.charAt(sIndex);
						if (c < '\u0080') {
							buf[i++] = (byte) c;
						} else if (c < '\u0800') {
							buf[i++] = (byte) (192 | c >>> 6);
							buf[i++] = (byte) (128 | c & 63);
						} else if (c < '\ud800' || c > '\udfff') {
							buf[i++] = (byte) (224 | c >>> 12);
							buf[i++] = (byte) (128 | c >>> 6 & 63);
							buf[i++] = (byte) (128 | c & 63);
						} else {
							int cp = 0;
							if (++sIndex < sLength) cp = Character.toCodePoint(c, s.charAt(sIndex));
							if ((cp >= 1 << 16) && (cp < 1 << 21)) {
								buf[i++] = (byte) (240 | cp >>> 18);
								buf[i++] = (byte) (128 | cp >>> 12 & 63);
								buf[i++] = (byte) (128 | cp >>> 6 & 63);
								buf[i++] = (byte) (128 | cp & 63);
							} else
								buf[i++] = (byte) '?';
						}
					}
					int size = i - start;
					if (size > {{$class}}.colferSizeMax)
						throw new IllegalStateException(format("colfer: {{.String}}[%d] size %d exceeds %d UTF-8 bytes", ai, size, {{$class}}.colferSizeMax));

					int ii = start - 1;
					if (size > 0x7f) {
						i++;
						for (int y = size; y >= 1 << 14; y >>>= 7) i++;
						System.arraycopy(buf, start, buf, i - size, size);

						do {
							buf[ii++] = (byte) (size | 0x80);
							size >>>= 7;
						} while (size > 0x7f);
					}
					buf[ii] = (byte) size;
				}
			}
 {{- else}}
			if (! this.{{.NameNative}}.isEmpty()) {
				buf[i++] = (byte) {{.Index}};
				int start = ++i;

				String s = this.{{.NameNative}};
				for (int sIndex = 0, sLength = s.length(); sIndex < sLength; sIndex++) {
					char c = s.charAt(sIndex);
					if (c < '\u0080') {
						buf[i++] = (byte) c;
					} else if (c < '\u0800') {
						buf[i++] = (byte) (192 | c >>> 6);
						buf[i++] = (byte) (128 | c & 63);
					} else if (c < '\ud800' || c > '\udfff') {
						buf[i++] = (byte) (224 | c >>> 12);
						buf[i++] = (byte) (128 | c >>> 6 & 63);
						buf[i++] = (byte) (128 | c & 63);
					} else {
						int cp = 0;
						if (++sIndex < sLength) cp = Character.toCodePoint(c, s.charAt(sIndex));
						if ((cp >= 1 << 16) && (cp < 1 << 21)) {
							buf[i++] = (byte) (240 | cp >>> 18);
							buf[i++] = (byte) (128 | cp >>> 12 & 63);
							buf[i++] = (byte) (128 | cp >>> 6 & 63);
							buf[i++] = (byte) (128 | cp & 63);
						} else
							buf[i++] = (byte) '?';
					}
				}
				int size = i - start;
				if (size > {{$class}}.colferSizeMax)
					throw new IllegalStateException(format("colfer: {{.String}} size %d exceeds %d UTF-8 bytes", size, {{$class}}.colferSizeMax));

				int ii = start - 1;
				if (size > 0x7f) {
					i++;
					for (int x = size; x >= 1 << 14; x >>>= 7) i++;
					System.arraycopy(buf, start, buf, i - size, size);

					do {
						buf[ii++] = (byte) (size | 0x80);
						size >>>= 7;
					} while (size > 0x7f);
				}
				buf[ii] = (byte) size;
			}
 {{- end}}
{{else if eq .Type "binary"}}
 {{- if .TypeList}}
			if (this.{{.NameNative}}.length != 0) {
				buf[i++] = (byte) {{.Index}};
				byte[][] a = this.{{.NameNative}};

				int x = a.length;
				if (x > {{$class}}.colferListMax)
					throw new IllegalStateException(format("colfer: {{.String}} length %d exceeds %d elements", x, {{$class}}.colferListMax));
				while (x > 0x7f) {
					buf[i++] = (byte) (x | 0x80);
					x >>>= 7;
				}
				buf[i++] = (byte) x;

				for (int ai = 0; ai < a.length; ai++) {
					byte[] b = a[ai];
					if (b == null) {
						b = _zeroBytes;
						a[ai] = b;
					}
					if (b.length > {{$class}}.colferSizeMax)
						throw new IllegalStateException(format("colfer: {{.String}}[%d] size %d exceeds %d bytes", ai, b.length, {{$class}}.colferSizeMax));

					x = b.length;
					while (x > 0x7f) {
						buf[i++] = (byte) (x | 0x80);
						x >>>= 7;
					}
					buf[i++] = (byte) x;

					int start = i;
					i += b.length;
					System.arraycopy(b, 0, buf, start, b.length);
				}
			}
 {{- else}}
			if (this.{{.NameNative}}.length != 0) {
				buf[i++] = (byte) {{.Index}};

				int size = this.{{.NameNative}}.length;
				if (size > {{$class}}.colferSizeMax)
					throw new IllegalStateException(format("colfer: {{.String}} size %d exceeds %d bytes", size, {{$class}}.colferSizeMax));

				int x = size;
				while (x > 0x7f) {
					buf[i++] = (byte) (x | 0x80);
					x >>>= 7;
				}
				buf[i++] = (byte) x;

				int start = i;
				i += size;
				System.arraycopy(this.{{.NameNative}}, 0, buf, start, size);
			}
 {{- end}}
{{else if .TypeList}}
			if (this.{{.NameNative}}.length != 0) {
				buf[i++] = (byte) {{.Index}};
				{{.TypeNative}}[] a = this.{{.NameNative}};

				int x = a.length;
				if (x > {{$class}}.colferListMax)
					throw new IllegalStateException(format("colfer: {{.String}} length %d exceeds %d elements", x, {{$class}}.colferListMax));
				while (x > 0x7f) {
					buf[i++] = (byte) (x | 0x80);
					x >>>= 7;
				}
				buf[i++] = (byte) x;

				for (int ai = 0; ai < a.length; ai++) {
					{{.TypeNative}} o = a[ai];
					if (o == null) {
						o = new {{.TypeNative}}();
						a[ai] = o;
					}
					i = o.marshal(buf, i);
				}
			}
{{else}}
			if (this.{{.NameNative}} != null) {
				buf[i++] = (byte) {{.Index}};
				i = this.{{.NameNative}}.marshal(buf, i);
			}
{{end}}{{end}}
			buf[i++] = (byte) 0x7f;
			return i;
		} catch (ArrayIndexOutOfBoundsException e) {
			if (i - offset > {{$class}}.colferSizeMax)
				throw new IllegalStateException(format("colfer: {{.String}} exceeds %d bytes", {{$class}}.colferSizeMax));
			if (i > buf.length) throw new BufferOverflowException();
			throw e;
		}
	}

	/**
	 * Deserializes the object.
	 * @param buf the data source.
	 * @param offset the initial index for {@code buf}, inclusive.
	 * @return the final index for {@code buf}, exclusive.
	 * @throws BufferUnderflowException when {@code buf} is incomplete. (EOF)
	 * @throws SecurityException on an upper limit breach defined by{{if .HasList}} either{{end}} {@link #colferSizeMax}{{if .HasList}} or {@link #colferListMax}{{end}}.
	 * @throws InputMismatchException when the data does not match this object's schema.
	 */
	public int unmarshal(byte[] buf, int offset) {
		return unmarshal(buf, offset, buf.length);
	}

	/**
	 * Deserializes the object.
	 * @param buf the data source.
	 * @param offset the initial index for {@code buf}, inclusive.
	 * @param end the index limit for {@code buf}, exclusive.
	 * @return the final index for {@code buf}, exclusive.
	 * @throws BufferUnderflowException when {@code buf} is incomplete. (EOF)
	 * @throws SecurityException on an upper limit breach defined by{{if .HasList}} either{{end}} {@link #colferSizeMax}{{if .HasList}} or {@link #colferListMax}{{end}}.
	 * @throws InputMismatchException when the data does not match this object's schema.
	 */
	public int unmarshal(byte[] buf, int offset, int end) {
		if (end > buf.length) end = buf.length;
		int i = offset;

		try {
			byte header = buf[i++];
{{range .Fields}}{{if eq .Type "bool"}}
			if (header == (byte) {{.Index}}) {
				this.{{.NameNative}} = true;
				header = buf[i++];
			}
{{else if eq .Type "uint8"}}
			if (header == (byte) {{.Index}}) {
				this.{{.NameNative}} = buf[i++];
				header = buf[i++];
			}
{{else if eq .Type "uint16"}}
			if (header == (byte) {{.Index}}) {
				this.{{.NameNative}} = (short) ((buf[i++] & 0xff) << 8 | (buf[i++] & 0xff));
				header = buf[i++];
			} else if (header == (byte) ({{.Index}} | 0x80)) {
				this.{{.NameNative}} = (short) (buf[i++] & 0xff);
				header = buf[i++];
			}
{{else if eq .Type "uint32"}}
			if (header == (byte) {{.Index}}) {
				int x = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					x |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				this.{{.NameNative}} = x;
				header = buf[i++];
			} else if (header == (byte) ({{.Index}} | 0x80)) {
				this.{{.NameNative}} = (buf[i++] & 0xff) << 24 | (buf[i++] & 0xff) << 16 | (buf[i++] & 0xff) << 8 | (buf[i++] & 0xff);
				header = buf[i++];
			}
{{else if eq .Type "uint64"}}
			if (header == (byte) {{.Index}}) {
				long x = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					if (shift == 56 || b >= 0) {
						x |= (b & 0xffL) << shift;
						break;
					}
					x |= (b & 0x7fL) << shift;
				}
				this.{{.NameNative}} = x;
				header = buf[i++];
			} else if (header == (byte) ({{.Index}} | 0x80)) {
				this.{{.NameNative}} = (buf[i++] & 0xffL) << 56 | (buf[i++] & 0xffL) << 48 | (buf[i++] & 0xffL) << 40 | (buf[i++] & 0xffL) << 32
					| (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
				header = buf[i++];
			}
{{else if eq .Type "int32"}}
			if (header == (byte) {{.Index}}) {
				int x = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					x |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				this.{{.NameNative}} = x;
				header = buf[i++];
			} else if (header == (byte) ({{.Index}} | 0x80)) {
				int x = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					x |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				this.{{.NameNative}} = -x;
				header = buf[i++];
			}
{{else if eq .Type "int64"}}
			if (header == (byte) {{.Index}}) {
				long x = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					if (shift == 56 || b >= 0) {
						x |= (b & 0xffL) << shift;
						break;
					}
					x |= (b & 0x7fL) << shift;
				}
				this.{{.NameNative}} = x;
				header = buf[i++];
			} else if (header == (byte) ({{.Index}} | 0x80)) {
				long x = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					if (shift == 56 || b >= 0) {
						x |= (b & 0xffL) << shift;
						break;
					}
					x |= (b & 0x7fL) << shift;
				}
				this.{{.NameNative}} = -x;
				header = buf[i++];
			}
{{else if eq .Type "float32"}}
			if (header == (byte) {{.Index}}) {
 {{- if .TypeList}}
				int length = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					length |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (length < 0 || length > {{$class}}.colferListMax)
					throw new SecurityException(format("colfer: {{.String}} length %d exceeds %d elements", length, {{$class}}.colferListMax));

				float[] a = new float[length];
				for (int ai = 0; ai < length; ai++) {
					int x = (buf[i++] & 0xff) << 24 | (buf[i++] & 0xff) << 16 | (buf[i++] & 0xff) << 8 | (buf[i++] & 0xff);
					a[ai] = Float.intBitsToFloat(x);
				}
				this.{{.NameNative}} = a;
 {{- else}}
				int x = (buf[i++] & 0xff) << 24 | (buf[i++] & 0xff) << 16 | (buf[i++] & 0xff) << 8 | (buf[i++] & 0xff);
				this.{{.NameNative}} = Float.intBitsToFloat(x);
 {{- end}}
				header = buf[i++];
			}
{{else if eq .Type "float64"}}
			if (header == (byte) {{.Index}}) {
 {{- if .TypeList}}
				int length = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					length |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (length < 0 || length > {{$class}}.colferListMax)
					throw new SecurityException(format("colfer: {{.String}} length %d exceeds %d elements", length, {{$class}}.colferListMax));

				double[] a = new double[length];
				for (int ai = 0; ai < length; ai++) {
					long x = (buf[i++] & 0xffL) << 56 | (buf[i++] & 0xffL) << 48 | (buf[i++] & 0xffL) << 40 | (buf[i++] & 0xffL) << 32
						| (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
					a[ai] = Double.longBitsToDouble(x);
				}
				this.{{.NameNative}} = a;
 {{- else}}
				long x = (buf[i++] & 0xffL) << 56 | (buf[i++] & 0xffL) << 48 | (buf[i++] & 0xffL) << 40 | (buf[i++] & 0xffL) << 32
					| (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
				this.{{.NameNative}} = Double.longBitsToDouble(x);
 {{- end}}
				header = buf[i++];
			}
{{else if eq .Type "timestamp"}}
			if (header == (byte) {{.Index}}) {
				long s = (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
				long ns = (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
				this.{{.NameNative}} = java.time.Instant.ofEpochSecond(s, ns);
				header = buf[i++];
			} else if (header == (byte) ({{.Index}} | 0x80)) {
				long s = (buf[i++] & 0xffL) << 56 | (buf[i++] & 0xffL) << 48 | (buf[i++] & 0xffL) << 40 | (buf[i++] & 0xffL) << 32
					| (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
				long ns = (buf[i++] & 0xffL) << 24 | (buf[i++] & 0xffL) << 16 | (buf[i++] & 0xffL) << 8 | (buf[i++] & 0xffL);
				this.{{.NameNative}} = java.time.Instant.ofEpochSecond(s, ns);
				header = buf[i++];
			}
{{else if eq .Type "text"}}
			if (header == (byte) {{.Index}}) {
 {{- if .TypeList}}
				int length = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					length |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (length < 0 || length > {{$class}}.colferListMax)
					throw new SecurityException(format("colfer: {{.String}} length %d exceeds %d elements", length, {{$class}}.colferListMax));

				{{.TypeNative}}[] a = new {{.TypeNative}}[length];
				for (int ai = 0; ai < length; ai++) {
					int size = 0;
					for (int shift = 0; true; shift += 7) {
						byte b = buf[i++];
						size |= (b & 0x7f) << shift;
						if (shift == 28 || b >= 0) break;
					}
					if (size < 0 || size > {{$class}}.colferSizeMax)
						throw new SecurityException(format("colfer: {{.String}}[%d] size %d exceeds %d UTF-8 bytes", ai, size, {{$class}}.colferSizeMax));

					int start = i;
					i += size;
					a[ai] = new String(buf, start, size, StandardCharsets.UTF_8);
				}
				this.{{.NameNative}} = a;
 {{- else}}
				int size = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					size |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (size < 0 || size > {{$class}}.colferSizeMax)
					throw new SecurityException(format("colfer: {{.String}} size %d exceeds %d UTF-8 bytes", size, {{$class}}.colferSizeMax));

				int start = i;
				i += size;
				this.{{.NameNative}} = new String(buf, start, size, StandardCharsets.UTF_8);
 {{- end}}
				header = buf[i++];
			}
{{else if eq .Type "binary"}}
 {{- if .TypeList}}
			if (header == (byte) {{.Index}}) {
				int length = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					length |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (length < 0 || length > {{$class}}.colferListMax)
					throw new SecurityException(format("colfer: {{.String}} length %d exceeds %d elements", length, {{$class}}.colferListMax));

				byte[][] a = new byte[length][];
				for (int ai = 0; ai < length; ai++) {
					int size = 0;
					for (int shift = 0; true; shift += 7) {
						byte b = buf[i++];
						size |= (b & 0x7f) << shift;
						if (shift == 28 || b >= 0) break;
					}
					if (size < 0 || size > {{$class}}.colferSizeMax)
						throw new SecurityException(format("colfer: {{.String}}[%d] size %d exceeds %d bytes", ai, size, {{$class}}.colferSizeMax));

					byte[] e = new byte[size];
					int start = i;
					i += size;
					System.arraycopy(buf, start, e, 0, size);
					a[ai] = e;
				}
				this.{{.NameNative}} = a;

				header = buf[i++];
			}
 {{- else}}
			if (header == (byte) {{.Index}}) {
				int size = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					size |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (size < 0 || size > {{$class}}.colferSizeMax)
					throw new SecurityException(format("colfer: {{.String}} size %d exceeds %d bytes", size, {{$class}}.colferSizeMax));

				this.{{.NameNative}} = new byte[size];
				int start = i;
				i += size;
				System.arraycopy(buf, start, this.{{.NameNative}}, 0, size);

				header = buf[i++];
			}
 {{- end}}
{{else if .TypeList}}
			if (header == (byte) {{.Index}}) {
				int length = 0;
				for (int shift = 0; true; shift += 7) {
					byte b = buf[i++];
					length |= (b & 0x7f) << shift;
					if (shift == 28 || b >= 0) break;
				}
				if (length < 0 || length > {{$class}}.colferListMax)
					throw new SecurityException(format("colfer: {{.String}} length %d exceeds %d elements", length, {{$class}}.colferListMax));

				{{.TypeNative}}[] a = new {{.TypeNative}}[length];
				for (int ai = 0; ai < length; ai++) {
					{{.TypeNative}} o = new {{.TypeNative}}();
					i = o.unmarshal(buf, i, end);
					a[ai] = o;
				}
				this.{{.NameNative}} = a;
				header = buf[i++];
			}
{{else}}
			if (header == (byte) {{.Index}}) {
				this.{{.NameNative}} = new {{.TypeNative}}();
				i = this.{{.NameNative}}.unmarshal(buf, i, end);
				header = buf[i++];
			}
{{end}}{{end}}
			if (header != (byte) 0x7f)
				throw new InputMismatchException(format("colfer: unknown header at byte %d", i - 1));
		} finally {
			if (i > end && end - offset < {{$class}}.colferSizeMax) throw new BufferUnderflowException();
			if (i < 0 || i - offset > {{$class}}.colferSizeMax)
				throw new SecurityException(format("colfer: {{.String}} exceeds %d bytes", {{$class}}.colferSizeMax));
			if (i > end) throw new BufferUnderflowException();
		}

		return i;
	}

	// {@link Serializable} version number.
	private static final long serialVersionUID = {{len .Fields}}L;

	// {@link Serializable} Colfer extension.
	private void writeObject(ObjectOutputStream out) throws IOException {
		// TODO: better size estimation
		byte[] buf = new byte[1024];
		int n;
		while (true) try {
			n = marshal(buf, 0);
			break;
		} catch (BufferUnderflowException e) {
			buf = new byte[4 * buf.length];
		}

		out.writeInt(n);
		out.write(buf, 0, n);
	}

	// {@link Serializable} Colfer extension.
	private void readObject(ObjectInputStream in) throws ClassNotFoundException, IOException {
		init();

		int n = in.readInt();
		byte[] buf = new byte[n];
		in.readFully(buf);
		unmarshal(buf, 0);
	}

	// {@link Serializable} Colfer extension.
	private void readObjectNoData() throws ObjectStreamException {
		init();
	}
{{range .Fields}}
	/**
	 * Gets {{.String}}.
	 * @return the value.
	 */
	public {{.TypeNative}}{{if .TypeList}}[]{{end}} get{{.NameTitle}}() {
		return this.{{.NameNative}};
	}

	/**
	 * Sets {{.String}}.
	 * @param value the replacement.
	 */
	public void set{{.NameTitle}}({{.TypeNative}}{{if .TypeList}}[]{{end}} value) {
		this.{{.NameNative}} = value;
	}

	/**
	 * Sets {{.String}}.
	 * @param value the replacement.
	 * @return {@code this}.
	 */
	public {{$class}} with{{.NameTitle}}({{.TypeNative}}{{if .TypeList}}[]{{end}} value) {
		this.{{.NameNative}} = value;
		return this;
	}
{{end}}
	@Override
	public final int hashCode() {
		int h = 1;
{{- range .Fields}}
{{- if eq .Type "bool"}}
		h = 31 * h + (this.{{.NameNative}} ? 1231 : 1237);
{{- else if eq .Type "uint8"}}
		h = 31 * h + (this.{{.NameNative}} & 0xff);
{{- else if eq .Type "uint16"}}
		h = 31 * h + (this.{{.NameNative}} & 0xffff);
{{- else if eq .Type "uint32" "int32"}}
		h = 31 * h + this.{{.NameNative}};
{{- else if eq .Type "uint64" "int64"}}
		h = 31 * h + (int)(this.{{.NameNative}} ^ this.{{.NameNative}} >>> 32);
{{- else if eq .Type "float32"}}
 {{- if .TypeList}}
		h = 31 * h + java.util.Arrays.hashCode(this.{{.NameNative}});
 {{- else}}
		h = 31 * h + Float.floatToIntBits(this.{{.NameNative}});
 {{- end}}
{{- else if eq .Type "float64"}}
 {{- if .TypeList}}
		h = 31 * h + java.util.Arrays.hashCode(this.{{.NameNative}});
 {{- else}}
		long _{{.NameNative}}Bits = Double.doubleToLongBits(this.{{.NameNative}});
		h = 31 * h + (int) (_{{.NameNative}}Bits ^ _{{.NameNative}}Bits >>> 32);
 {{- end}}
{{- else if eq .Type "binary"}}
 {{- if .TypeList}}
		for (byte[] b : this.{{.NameNative}}) h = 31 * h + java.util.Arrays.hashCode(b);
 {{- else}}
		for (byte b : this.{{.NameNative}}) h = 31 * h + b;
 {{- end}}
{{- else if .TypeList}}
		for ({{.TypeNative}} o : this.{{.NameNative}}) h = 31 * h + (o == null ? 0 : o.hashCode());
{{- else}}
		if (this.{{.NameNative}} != null) h = 31 * h + this.{{.NameNative}}.hashCode();
{{- end}}{{end}}
		return h;
	}

	@Override
	public final boolean equals(Object o) {
		return o instanceof {{$class}} && equals(({{$class}}) o);
	}

	public final boolean equals({{$class}} o) {
		if (o == null) return false;
		if (o == this) return true;
		return o.getClass() == {{$class}}.class
{{- range .Fields}}
{{- if .TypeList}}
 {{- if eq .Type "binary"}}
			&& _equals(this.{{.NameNative}}, o.{{.NameNative}})
 {{- else}}
			&& java.util.Arrays.equals(this.{{.NameNative}}, o.{{.NameNative}})
 {{- end}}
{{- else if eq .Type "bool" "uint8" "uint16" "uint32" "uint64" "int32" "int64"}}
			&& this.{{.NameNative}} == o.{{.NameNative}}
{{- else if eq .Type "float32" "float64"}}
			&& (this.{{.NameNative}} == o.{{.NameNative}} || (this.{{.NameNative}} != this.{{.NameNative}} && o.{{.NameNative}} != o.{{.NameNative}}))
{{- else if eq .Type "binary"}}
			&& java.util.Arrays.equals(this.{{.NameNative}}, o.{{.NameNative}})
{{- else}}
			&& (this.{{.NameNative}} == null ? o.{{.NameNative}} == null : this.{{.NameNative}}.equals(o.{{.NameNative}}))
{{- end}}{{end}};
	}
{{if .HasBinaryList}}
	private static boolean _equals(byte[][] a, byte[][] b) {
		if (a == b) return true;
		if (a == null || b == null) return false;

		int i = a.length;
		if (i != b.length) return false;

		while (--i >= 0) if (! java.util.Arrays.equals(a[i], b[i])) return false;
		return true;
	}
{{end}}
}
`
