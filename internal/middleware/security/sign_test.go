package security

import "testing"

func TestCheckHash(t *testing.T) {
	type args struct {
		correctHash string
		requestHash string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test not equal",
			args: struct {
				correctHash string
				requestHash string
			}{
				correctHash: "43baa3a4500d582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
				requestHash: "4qwrrg582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
			},
			wantErr: true,
		},
		{
			name: "Test equal",
			args: struct {
				correctHash string
				requestHash string
			}{
				correctHash: "43baa3a4500d582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
				requestHash: "43baa3a4500d582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := CheckHash(
					tt.args.correctHash,
					tt.args.requestHash,
				); (err != nil) != tt.wantErr {
					t.Errorf("CheckHash() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
