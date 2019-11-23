package tls

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCertificate_GenerateRSAKey(t *testing.T) {
	emptyKP := NewCertificate("empty", nil)
	emptyKPwithErr := NewCertificate("empty", nil).withErrf("some error here")

	tests := []struct {
		name    string
		crt     *Certificate
		bits    int
		wantErr bool
	}{
		{"size 0/default", emptyKP, 0, false},
		{"size 1024", emptyKP, 1024, false},
		{"size 2048", emptyKP, 2048, false},
		{"size 4096", emptyKP, 4096, false},
		{"with error", emptyKPwithErr, 0, true},
		{"small size", emptyKP, 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.crt.WithBits(tt.bits).GenerateRSAKey().Error(); (err != nil) != tt.wantErr {
				t.Errorf("Certificate.GenerateRSAKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				validateCertificate(t, tt.crt, tt.bits, "GenerateRSAKey")
			}
		})
	}
}

func TestCertificate_GeneratePrivateKeyFromPEM(t *testing.T) {
	emptyKP := NewCertificate("empty", nil)
	emptyKPwithErr := NewCertificate("empty", nil).withErrf("some error here")

	type args struct {
		data       []byte
		passphrase string
	}
	tests := []struct {
		name    string
		crt     *Certificate
		size    int
		args    args
		wantErr bool
	}{
		{"from 1024 RSA key", emptyKP, 1024, args{[]byte(goodRSAPrivateKeyPEM1024), ""}, false},
		{"from 2048 RSA key", emptyKP, 2048, args{[]byte(goodRSAPrivateKeyPEM2048), ""}, false},
		{"from 4096 RSA key", emptyKP, 4096, args{[]byte(goodRSAPrivateKeyPEM4096), ""}, false},
		{"from generated 1024 RSA key", emptyKP, 1024, args{nil, ""}, false},
		{"from generated 2048 RSA key", emptyKP, 2048, args{nil, ""}, false},
		{"from generated 4096 RSA key", emptyKP, 4096, args{nil, ""}, false},
		{"from unencrypted 1024 RSA key", emptyKP, 1024, args{[]byte(goodRSAPrivateKeyPEM1024), "fake"}, false},
		{"from encrypted 1024 RSA key", emptyKP, 1024, args{[]byte(goodRSAPrivateKeyPEM1024Encrypted), "TesT1ng"}, false},
		{"wrong passphrase", emptyKP, 1024, args{[]byte(goodRSAPrivateKeyPEM1024Encrypted), "wrongpasswd"}, true},
		{"bad RSA key", emptyKP, 0, args{[]byte(badRSAPrivateKeyPEM01), ""}, true},
		{"bad RSA key type", emptyKP, 0, args{[]byte(badRSAPrivateKeyPEM02), ""}, true},
		{"bad RSA key end", emptyKP, 0, args{[]byte(badRSAPrivateKeyPEM03), ""}, true},
		{"with error", emptyKPwithErr, 0, args{[]byte("empty"), ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := tt.args.data
			if data == nil {
				crt := NewCertificate("empty", nil).WithBits(tt.size).GenerateRSAKey()
				data = crt.PrivateKeyPEM()
				if data == nil {
					t.Errorf("Certificate.PrivateKeyPEM() error %v", crt.Error())
				}
			}

			tt.crt.WithPassphrase(tt.args.passphrase).GeneratePrivateKeyFromPEM(data)
			if err := tt.crt.Error(); (err != nil) != tt.wantErr {
				t.Errorf("Certificate.GeneratePrivateKeyFromPEM() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				validateCertificate(t, tt.crt, tt.size, "GeneratePrivateKeyFromPEM")
			}
		})
	}
}

func TestCertificate_ReadPrivateKeyFromFile(t *testing.T) {
	emptyKP := NewCertificate("empty", nil)
	emptyKPwithErr := NewCertificate("empty", nil).withErrf("some error here")

	type args struct {
		filename   string
		passphrase string
	}
	tests := []struct {
		name    string
		crt     *Certificate
		content []byte
		args    args
		size    int
		wantErr bool
	}{
		{"generate with size 0/default", emptyKP, nil, args{"gendefault", ""}, 0, false},
		{"generate with size 1024", emptyKP, nil, args{"gen1024", ""}, 1024, false},
		{"generate with size 2048", emptyKP, nil, args{"gen2048", ""}, 2048, false},
		{"generate with size 4096", emptyKP, nil, args{"gen4096", ""}, 4096, false},
		{"from 1024 RSA key", emptyKP, []byte(goodRSAPrivateKeyPEM1024), args{"from1024", ""}, 1024, false},
		{"from 2048 RSA key", emptyKP, []byte(goodRSAPrivateKeyPEM2048), args{"from2048", ""}, 2048, false},
		{"from 4096 RSA key", emptyKP, []byte(goodRSAPrivateKeyPEM4096), args{"from4096", ""}, 4096, false},
		{"from unencrypted 1024 RSA key", emptyKP, []byte(goodRSAPrivateKeyPEM1024), args{"unc1024", "fake"}, 1024, false},
		{"from encrypted 1024 RSA key", emptyKP, []byte(goodRSAPrivateKeyPEM1024Encrypted), args{"enc1024", "TesT1ng"}, 1024, false},
		{"wrong passphrase", emptyKP, []byte(goodRSAPrivateKeyPEM1024Encrypted), args{"wrong", "wrongpasswd"}, 1024, true},
		{"bad RSA key", emptyKP, []byte(badRSAPrivateKeyPEM01), args{"bad01", ""}, 0, true},
		{"bad RSA key type", emptyKP, []byte(badRSAPrivateKeyPEM02), args{"bad02", ""}, 0, true},
		{"bad RSA key end", emptyKP, []byte(badRSAPrivateKeyPEM03), args{"bad03", ""}, 0, true},
		{"with error", emptyKPwithErr, nil, args{"err", ""}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpfile, err := ioutil.TempFile("", tt.args.filename)
			if err != nil {
				t.Errorf("Certificate.GenerateFromPrivateKeyFile() failed to generate temporal filename. %v", err)
			}
			filename := tmpfile.Name()
			defer os.Remove(filename)

			// if no content, generate an RSA key. If content, use it to create the file
			if tt.content == nil {
				err := NewCertificate("empty", nil).WithBits(tt.size).GenerateRSAKey().WritePrivateKeyToFile(filename).Error()
				if err != nil {
					t.Errorf("Certificate.WritePrivateKeyToFile() error = %v", err)
				}
			} else {
				err := ioutil.WriteFile(filename, tt.content, 0600)
				if err != nil {
					t.Errorf("Certificate.ReadPrivateKeyFromFile() failed to create temporal file. %v", err)
				}
			}

			// Generate the KP from the temporal file
			tt.crt.WithPassphrase(tt.args.passphrase).ReadPrivateKeyFromFile(filename)

			if err := tt.crt.Error(); (err != nil) != tt.wantErr {
				t.Errorf("Certificate.ReadPrivateKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				validateCertificate(t, tt.crt, tt.size, "ReadPrivateKeyFromFile")
			}
		})
	}
}

const (
	// generated with: `openssl genrsa -out test.key 1024`
	goodRSAPrivateKeyPEM1024 = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDdMT6j4GUEHjsGYneYYMBfd2omtUK0rFF79vuAhRTwWh1aMYJx
QDwsuNpiW56Tpd+xcTu0ULRK+ZqKkYq3u7RrMtIbTRrozx2LevXaaFf4oAALVe96
yUYpUuXCTeR31INyZ6RrSe5WxH7otup/q5ffk1sRzsT5HmgEDHo0CuJIVwIDAQAB
AoGAVU4XpWX2L2dlRioVi5vMtUS9mJ4RUtKePlvbVQ5+K8/lQBKWBw1J58mG16YT
+0dehPVxnDH/xmaLFmPZBtEKXBFuxmIRKnh7WDegxiGRhfbJ8fX1LbDI7c6vS4fI
zlILfqaLFDWwj49t9z1A5Hiit7d/gbkobfY/pvG/9UjvvXECQQD+okNftPI+Y3gN
QJtQDPm99VXvIRTSbFgm0mfhnl92MDtPSkDWuEeZcrNBc2gbEeN9YrGMoWrimsna
JnhWu7JvAkEA3mEMzNbmzHeVs+OY6WauijUo+EBWnBUSzJemjwgeCCcmGVZhVtVM
YBNc6ueiMGvj0MeTAYbvoiJ6SUTx9BecmQJAEBbWpPx89f44/rVfWu6G9T7EQEjZ
1PXtl/5he/sS0xf3F2p8sFSSiZqawv70zAkUU77hsx4b79eR+4vEvM49VwJBAI8l
9AJsF9fo5tIMoWMleRd4ju33JImGuzo+KLaL6CEhalIHG9O0rFxwOnDwOKRo3xaH
Uec6wEkjXoqJuiKBjIkCQQCdyaAB4EBVqu7q6/wEOQtRaUnjPmNJjSGBwtDGfr8Y
vJ7wYEg311CXghtjXjpuzdNmpJD9q65D360sNI6O6Tki
-----END RSA PRIVATE KEY-----`

	// generated with: `openssl genrsa -out test.key 2048`
	goodRSAPrivateKeyPEM2048 = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA220QzYWyZ5sAmMsvE9+3s9Q78Ym4xkDJklP1NUHeANmgja4m
b04leet6PHrzloCHN9uiW5JOFIBAiIo7EnQIIUTTl+gryHzag0zTqYnMeutBhE2k
48eit6UTjfRPggJ0j9rA+zNe3S6xo3KR427usKrCWmcsT8gi9a6sI5zPpYWbzalj
btYP7BIWvItZsly4AvsxFtRz0DdItaTSjCUJO63WyAPl/nYMQoMyFDOTMV4NsldH
MdcsEK8qgzutx0+wwof4T0MQ+uNGjrXLy9SaeJkKMIWWPuunq4soiAdhE+/xbV1t
w8+D7KP44igw9liyN+vTi+dBe9rgsBBzWVGkMwIDAQABAoIBAGJf5XafCDsW7Q+T
mImqdSu0bSN/AU4w4a4u4fJ8k1GCQE8E0sp4GKkuf2D6JpdFDFpop8OyNjtOvJ/6
iT1es/5mYw7p0idSgoz3NrX9x6UcMAVm4hQU/siiw7+S1sqM20Fba8WzZEZ5lr6W
sP1sUKr4A6aQNveV3MaVByv4M8ga6AyeNlJlAPkzBDfMXbITskOfHKHNjF2HKTmW
CeBN1I/x+7IWLfKND7v6lgppKNnuW+TZsTRYHA5XOqZsBkKXc3ZfZ9qKPPdwjGNd
9W9q05RK4qYSr+EqSgpQg66z429bVRyWlsiOCs4GHg2LZJsw9NYfxUjU4++s0cOP
8CdrxdECgYEA8gk0gGUCXpbMCIU83E7f8pBpGWRlplIBJxNFXrZYjU/j+upMlwDh
P6h9NUZ8iq8pSV4k5loQjCnvoI77sB79R1hR0YRpzmq8JUhG8Y8s8P/fjIq6JemG
GfD1RszhXsJaVomytAX2C1LOAaU07Q0/HKH5ihF8Y83+LAZzYRalTqcCgYEA6BXr
TaSkYzrJakBLCz8x8VBVsW926+0cwMsjSVS/QB5EDO+hdN4VlvBmQ5PrzfLwfqsI
RwkCX6YsznCcYCjEz4NiTl3Q0BjPFGpubWJp++cKztw0C5jm8iXEBl3WgxnH5C6m
6jE8H4m1hMEuSNq1zg2EDr+haz11z+4rKZTe25UCgYA1RA9Y3mH6HSfEDEaLmZQx
ltR6cRzp0IxOZDBVGr48Q0PCGcRHSztHPjZn/h40iz4D1oM6asHPhv8kj7MWXIuT
Iv3GpWR/t/+2GMT8Lm1OkJFLJdC3vT/0/5GLRlWm/BDTkLRSaClw7oVF2WzfapXk
IEWBvD1q3Y+9cqH70PcKawKBgBtMwMIAVOnrN1gNOgaXp2tIVNwrTWTRHmm7O2pb
0qdbKAwRf/98RDPeEVlvo/Q02H895Rpd/+56YJjj/DD/eq8iOEUZmf9we6NeJaEu
S2M7OTU+B805bikbsiRBk4MWXEIGDtJLelQHYde81ZHyUCJtypPljLpNn3cW//LD
NTfFAoGBAMibn4EFQGYtSs6Qg2ycdVDwHiNpzljDyrEbsmbIVXljqkSx/AxqZ7pP
EEeCJOu8VjtWVbwENJ/imx8AjGZUdm7MWi0WvD4NwIoKw20oBN6iOGlMypC51oBf
FSij9yPTLvcS6TF+DJndIj7BK4JJopCdlzvXlxAvAy8cehq4JI6W
-----END RSA PRIVATE KEY-----`

	// generated with `openssl genrsa -out test.key 4096`
	goodRSAPrivateKeyPEM4096 = `-----BEGIN RSA PRIVATE KEY-----
MIIJKAIBAAKCAgEAmSWDwwofV4qwRR6xpaD7WItFBKu1Uf4KdvleJ608QevW9bo3
WxEygIQcS1mKzyuOwWvTt0Ib0K+Bv/+VuGWv9ZekOIy7EFOMjKu82KIw2VJCfT2O
WDNJA7GsqhLLw8jFG3iYlE/h6MBUy/LIBDQh8JgwxnCnXNt1oenXFiibNIBvXFit
kZeEqAg3PTsa5sRVpnGtBHaZqdV/xSiIUsGtG076UJQLNE5PZiGifj0kISDMkl3S
yE2Xgrx8yJMYOjX9pTHzi4pP/4HMg9gS4vFBKI1ceuvWhgjbqbzkJTUBMcTKZVdM
S4xJhUhHX3ivOwsQkM8MVUVrVLhqo7Hp7whtDWQJsHaeh5+now4H+dooTKolfsHi
nkgFilOgmShVCOYoeTS2AGXOWIHga+5AJpyDB3IYRkeC9HkvXKVSyvb3nnnvj69x
GTNF+QXODwywkuV3lRPsClar23pdJ6Xbt41gc5/1GRA3bDcMS/PgrzT97C3l9dxB
1qLXrXWgRXDa9krd3QHlnZm419gP2SksXsdLNzVuU07/e167xTX5TKDfVH31hvJu
/Z8650cVmo3R+09eG2HSARGdOtS3w42jIgk8KzX+pAZ2ahNZUSdPR7+LyHE1QVir
1EaUfBR8akNG0LBRgprULiC2nA42WVPOu/CSIi5o9xF++4qOgb8U1dVCahkCAwEA
AQKCAgBOnLQzJPoLrNRo1qWn+KesvuixyhZsPTRP+Bd91G4PBKm2P1108LH0yFvh
zdf0Gk7QOLJX6qZui/jVfrPzELevIKUaphrL+ac2+l0Y8sCSfahFs5hi7Ah5zfVs
4/MWz/Nl85SW5R54pGmiu0Ay8DPP1b+LzX4Lq1eJwog3BqcM3zTKuXZ3OsjMnFMG
qlVXycaRht6iYOLmDALnSLQZjvN2Bid1v1i1A9G8Al3v4SCvpK9+Ho05ZqL0aB7B
ssgH8Xj/GFIE+f2wkYdS0lJ3jhG9aDrlmjPnX5qqWZzfAINZccPXG2W9jgibUwvY
FRIWA8sAGlghWV4MAhIzbvs/SRr4rlobTD3TzCxWVCn7fAkFN7LJcv+kqlbcZo6q
T7Y769o6TbDMAEoBOxYdhI2MZQ9QjEirDolFRcPAYGJutOLY63RgCZgEa0s/6zed
Ivu+EBVINN5Y+CKZqinjdIC8T9JCpTQq+2nXILYrf+Qb+IIixGwRShPUjk1iGYnc
lxXq/N+243rA7yvSLFjRw7hxa5AXr1zwwf1+eEpKLQUfa0JYfTVS930tXYddKu44
EJISKSxQswxP7L9aHeTgomPVzKHjzqceBL5IImn67qcxHAPS9QLTRz4CYA4owUme
d5CpczS/utHGpTioNoskK2u/nD7l6wFNVeHRkMOvr6LtAzymxQKCAQEAy62i+AdI
xAqQ2agfsHnxlbVziTxo7W51pk3Ca/rtty38kIfRNrKLCFQksEtfWUCW92yjIi80
s8v1q9xonxwu1vgWPjFJBaOHUGmLwVFvUFlvScphagkGHqMpAbGcb0m54in1l1s/
ZOe22z7z4S0+VY492QLvVsoJSgyE7jdXk6BnwLR+JNQuFL+NxgNrsTFaCuKx4lWQ
ai4HW8jRjTmJ2+b4Cf+I8KCIJ+aF+3wYMRkVtND107ArlJdACasLIbnL9/QPoPQL
DFQ3Bqs2/Vp7vgpst+1NWr98kitmhc1MHIy7ufxcCtZRZ4hmyA8QqLysEO4Cqxs4
Hr1N0qrl1OV6+wKCAQEAwHzK5hkLgdcViKpW100zNPJQcPnbg13rEE2iferWyz5p
EdQl4Gp/s8xNXtJ9OveT2ETVnLjJDgAd2enwDLyr2a+uHg0jwvK9op2D0KCJdKjV
VJY+HiAo/cuHzeMpWeWrYsZKK8JhqKrc1zDXjByCHXqFO0xeMte1NYX+GLUXaH8b
98SctM1GlWlgGsgs1jaS3Gs6LvBlWjL4BNrj85rOEYS6mNYPL25Eh/PBM19Lfeuz
IzbKt+DBcFW7faomGuZ9hmsYxul0Nx7QFGdHzlaYY0gcWXI6E81SPyCE+6FW8zlX
/AlXaJOkjSF9F7Fevd1vZ3IMdeuIiEp/vqTgtlCi+wKCAQEAmvdOglXQeE/tq5zu
F3CAbb5z3FuZHeUoIFMTdOKjUPbtLe4YMPyhKcITdAq0zgyFkFUYvZ6bA51QGuWE
uGJSBP8WtVT8UQz9nCHh0kEqZ0DUmpkfivS4skDDp2VCN64pfrkcAX/MePKPvrpk
BdRNk/y4c8922Fq5vJxP24tB17F4nzb0rwK82M5xiNH7cwKwlo2qeCFP0mmY8a2W
OlQn3qcZ20mQIDyTu3/6OsBNC7YhPMSr9NcaIWD+uRxpSMy7Mrl/1p83dLbycuF7
4kGg2pPF6h7j9wKwPepFg9IScbpl7njicuBjaVlvkhFcRHXgmLTuVM9N4J15g6rB
WT0MVwKCAQBoZ4e6FJ2MOHhin1no/+OldUUyciOhdmCYgDOBrs4AgYKF/BbrSXio
skjJBMyOHllftb5Telr7MA8A8oWUswVXVpXPkPrzs6wuNteXYmwMDcNgVPmuZ200
c5/eibcVHqC+O6VhZNaJoNuWENTpF1Fv4dPAHST+2MdeF6aCCj9/G2q1EjyZSLpf
Mj/BZxACxKkVy0dMzHF84iZqo7t+l1nsYJzBZ2HnLR0YzJrfXXHaA+0vTXnZEJx7
bNT8TTzL3Hb7YL2YrmmtuBXO61IkVg3j0+okjfN4aCaTPPVEcvdxh4n0l7CEdYiM
UDzEjB1CSIgziMW/dBijLB1r74w+9y/fAoIBAAwvCB5/CoHNo+ewRFoya2f6k7Nx
sHl2SaNk4+7oW1FlMo8doF7jXpvRPLtkxq/SEzWLXa1NZzdOOiexHlPc3dgjrxOK
1dEmj4Q14m/oT2JKH98nfc3R7C5mfVLRFFGC3AiiAVmCr2OoiPWW2Kvjn76wla/2
xx6IvbQByo2d8ovYNupgwgDBqEQfA6RcuT7xg3er/K1/vLk8qXiiltc+p5kT4zXz
fhcsVwZBDCWj2Ps6w2Y/ChB6QKZuqaETWiX4zQKdi1lJmmBw0++1RHa3H3N6xDUZ
qDatCCXNgZAN9wV/+b+W0M3vKhrx779CPBM0Sa+32Bj/pdd3I1LpFxyenAc=
-----END RSA PRIVATE KEY-----`

	// generated with `openssl genrsa -des3 -passout pass:TesT1ng -out test.key 1024`
	goodRSAPrivateKeyPEM1024Encrypted = `-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,287DD26C6CA8DB26

MIhGVMpTuDecHJGyBMrhsz9wiTRz6uAB0jw8xeSUH1x2ZuD8B8Sv+4M65zBbo72C
0KyU9WssG1fDUqMlG9rWE8BQbX/FKjuem1KgCAQiJ570ZcoWwigBiN+UzzehbK1M
r3vYi9nmriLyylmMLKm1cMuhwHEtC/0xi/wSi775G6f8DtN0sIhgxVyiYG+NyGGG
Vzc54ZmbetipT4MWO3w+cApBhF+nXPGFGhCKjQNEjyJVNQHGsSC7ZYlSo8a4gnA1
cAwN/utybbJL28CKbSQsaI2tJrOJprINjaZ0Ryp+y2O/kejUIW8TbqwxgUod0/ki
b3Bzt4coHI7yz4N24sWwy3fgKyoUMaMey6gxfoLu6HoCvF0GDZBs443xwa6vz8dL
OMo43DiuI/QEOSay39JblTwFMrhxMKEUWKsn96DRw1pOw8YSx9d8lg/mq3WE5SyO
4N1TzKPmMFQ/oRGjd1nxA2hpCXaO71Y1O5WGOJBDxBF/0YhhQMFxiKr1sjLECYj4
7LK207b1TKVhh55fvzjpanGj24qAUXsH3Ssc0l4PnhmFOAcaJsv6jFft2cTkMdDO
FdJQw6GWV/elu0+yIPDBxC6igtP56IjnF97Wyp36jrn268zIHhhd09U+KtGlsZfa
+Ix3JlIT4F/I8jEoSWmrve1hA/1WyAH/KL54rV3Tll/RM0dhMuhfVjwJwEAhwpZO
gexOt8V0kI5zamiesoAaqNkS5FfniScmLJvM8c/QnMh4fccj5yaccG/aDjRcApoH
3nOu+uufckNmhljhU1kSC5H2JfwOLDtZTNdLg+0oLgLnheh7lARqkA==
-----END RSA PRIVATE KEY-----`

	badRSAPrivateKeyPEM01 = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA220QzYWyZ5sAmMsvE9+3s9Q78Ym4xkDJklP1NUHeANmgja4m
b04leet6PHrzloCHN9uiW5JOFIBAiIo7EnQIIUTTl+gryHzag0zTqYnMeutBhE2k
48eit6UTjfRPggJ0j9rA+zNe3S6xo3KR427usKrCWmcsT8gi9a6sI5zPpYWbzalj
btYP7BIWvItZsly4AvsxFtRz0DdItaTSjCUJO63WyAPl/nYMQoMyFDOTMV4NsldH
MdcsEK8qgzutx0+wwof4T0MQ+uNGjrXLy9SaeJkKMIWWPuunq4soiAdhE+/xbV1t
w8+D7KP44igw9liyN+vTi+dBe9rgsBBzWVGkMwIDAQABAoIBAGJf5XafCDsW7Q+T
0qdbKAwRf/98RDPeEVlvo/Q02H895Rpd/+56YJjj/DD/eq8iOEUZmf9we6NeJaEu
S2M7OTU+B805bikbsiRBk4MWXEIGDtJLelQHYde81ZHyUCJtypPljLpNn3cW//LD
NTfFAoGBAMibn4EFQGYtSs6Qg2ycdVDwHiNpzljDyrEbsmbIVXljqkSx/AxqZ7pP
EEeCJOu8VjtWVbwENJ/imx8AjGZUdm7MWi0WvD4NwIoKw20oBN6iOGlMypC51oBf
FSij9yPTLvcS6TF+DJndIj7BK4JJopCdlzvXlxAvAy8cehq4JI6W
-----END RSA PRIVATE KEY-----`

	badRSAPrivateKeyPEM02 = `-----BEGIN WHAT PRIVATE KEY-----
MIICXQIBAAKBgQDdMT6j4GUEHjsGYneYYMBfd2omtUK0rFF79vuAhRTwWh1aMYJx
QDwsuNpiW56Tpd+xcTu0ULRK+ZqKkYq3u7RrMtIbTRrozx2LevXaaFf4oAALVe96
yUYpUuXCTeR31INyZ6RrSe5WxH7otup/q5ffk1sRzsT5HmgEDHo0CuJIVwIDAQAB
AoGAVU4XpWX2L2dlRioVi5vMtUS9mJ4RUtKePlvbVQ5+K8/lQBKWBw1J58mG16YT
+0dehPVxnDH/xmaLFmPZBtEKXBFuxmIRKnh7WDegxiGRhfbJ8fX1LbDI7c6vS4fI
zlILfqaLFDWwj49t9z1A5Hiit7d/gbkobfY/pvG/9UjvvXECQQD+okNftPI+Y3gN
QJtQDPm99VXvIRTSbFgm0mfhnl92MDtPSkDWuEeZcrNBc2gbEeN9YrGMoWrimsna
JnhWu7JvAkEA3mEMzNbmzHeVs+OY6WauijUo+EBWnBUSzJemjwgeCCcmGVZhVtVM
YBNc6ueiMGvj0MeTAYbvoiJ6SUTx9BecmQJAEBbWpPx89f44/rVfWu6G9T7EQEjZ
1PXtl/5he/sS0xf3F2p8sFSSiZqawv70zAkUU77hsx4b79eR+4vEvM49VwJBAI8l
9AJsF9fo5tIMoWMleRd4ju33JImGuzo+KLaL6CEhalIHG9O0rFxwOnDwOKRo3xaH
Uec6wEkjXoqJuiKBjIkCQQCdyaAB4EBVqu7q6/wEOQtRaUnjPmNJjSGBwtDGfr8Y
vJ7wYEg311CXghtjXjpuzdNmpJD9q65D360sNI6O6Tki
-----END RSA PRIVATE KEY-----`

	badRSAPrivateKeyPEM03 = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDdMT6j4GUEHjsGYneYYMBfd2omtUK0rFF79vuAhRTwWh1aMYJx
QDwsuNpiW56Tpd+xcTu0ULRK+ZqKkYq3u7RrMtIbTRrozx2LevXaaFf4oAALVe96
yUYpUuXCTeR31INyZ6RrSe5WxH7otup/q5ffk1sRzsT5HmgEDHo0CuJIVwIDAQAB
AoGAVU4XpWX2L2dlRioVi5vMtUS9mJ4RUtKePlvbVQ5+K8/lQBKWBw1J58mG16YT
+0dehPVxnDH/xmaLFmPZBtEKXBFuxmIRKnh7WDegxiGRhfbJ8fX1LbDI7c6vS4fI
zlILfqaLFDWwj49t9z1A5Hiit7d/gbkobfY/pvG/9UjvvXECQQD+okNftPI+Y3gN
QJtQDPm99VXvIRTSbFgm0mfhnl92MDtPSkDWuEeZcrNBc2gbEeN9YrGMoWrimsna
JnhWu7JvAkEA3mEMzNbmzHeVs+OY6WauijUo+EBWnBUSzJemjwgeCCcmGVZhVtVM
YBNc6ueiMGvj0MeTAYbvoiJ6SUTx9BecmQJAEBbWpPx89f44/rVfWu6G9T7EQEjZ
1PXtl/5he/sS0xf3F2p8sFSSiZqawv70zAkUU77hsx4b79eR+4vEvM49VwJBAI8l
9AJsF9fo5tIMoWMleRd4ju33JImGuzo+KLaL6CEhalIHG9O0rFxwOnDwOKRo3xaH
Uec6wEkjXoqJuiKBjIkCQQCdyaAB4EBVqu7q6/wEOQtRaUnjPmNJjSGBwtDGfr8Y
vJ7wYEg311CXghtjXjpuzdNmpJD9q65D360sNI6O6Tki
-----END-----`
)
