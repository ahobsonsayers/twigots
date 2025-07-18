package twigots

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNewProxy(t *testing.T) {
	type args struct {
		host     string
		port     int
		user     string
		password string
	}
	tests := []struct {
		name string
		args args
		want *Proxy
	}{
		{
			name: "Basic Proxy",
			args: args{
				host:     "localhost",
				port:     8080,
				user:     "user",
				password: "password",
			},
			want: &Proxy{
				Host:     "localhost",
				Port:     8080,
				User:     "user",
				Password: "password",
			},
		},
		{
			name: "Proxy without user and password",
			args: args{
				host: "localhost",
				port: 8080,
			},
			want: &Proxy{
				Host:     "localhost",
				Port:     8080,
				User:     "",
				Password: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProxy(
				tt.args.host,
				tt.args.port,
				tt.args.user,
				tt.args.password,
			)
			if err != nil {
				t.Errorf("NewProxy() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_String(t *testing.T) {
	type fields struct {
		Host     string
		Port     int
		User     string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Valid Proxy with User and Password",
			fields: fields{
				Host:     "localhost",
				Port:     8080,
				User:     "user",
				Password: "password",
			},
			want: "socks5://user:password@localhost:8080",
		},
		{
			name: "Valid Proxy without User and Password",
			fields: fields{
				Host:     "localhost",
				Port:     8080,
				User:     "",
				Password: "",
			},
			want: "socks5://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proxy{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				User:     tt.fields.User,
				Password: tt.fields.Password,
			}
			got, err := p.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("Proxy.String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Proxy.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_URL(t *testing.T) {
	type fields struct {
		Host     string
		Port     int
		User     string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *url.URL
		wantErr bool
	}{
		{
			name: "Valid Proxy with User and Password",
			fields: fields{
				Host:     "localhost",
				Port:     8080,
				User:     "user",
				Password: "password",
			},
			want: &url.URL{
				Scheme: "socks5",
				Host:   "localhost:8080",
				User:   url.UserPassword("user", "password"),
			},
		},
		{
			name: "Valid Proxy without User and Password",
			fields: fields{
				Host:     "localhost",
				Port:     8080,
				User:     "",
				Password: "",
			},
			want: &url.URL{
				Scheme: "socks5",
				Host:   "localhost:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proxy{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				User:     tt.fields.User,
				Password: tt.fields.Password,
			}
			got, err := p.URL()
			if (err != nil) != tt.wantErr {
				t.Errorf("Proxy.URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Proxy.URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateProxyList(t *testing.T) {
	type args struct {
		proxyHosts []string
		user       string
		password   string
	}
	tests := []struct {
		name    string
		args    args
		want    []Proxy
		wantErr bool
	}{
		{
			name: "Valid Proxy List",
			args: args{
				proxyHosts: []string{"socks5://localhost:8080", "socks5://example.com:9090"},
				user:       "user",
				password:   "password",
			},
			want: []Proxy{
				{
					Host:     "localhost",
					Port:     8080,
					User:     "user",
					Password: "password",
				},
				{
					Host:     "example.com",
					Port:     9090,
					User:     "user",
					Password: "password",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateProxyList(tt.args.proxyHosts, tt.args.user, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateProxyList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateProxyList() = %v, want %v", got, tt.want)
			}
		})
	}
}
