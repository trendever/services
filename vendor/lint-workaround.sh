# go-vim (actually, it's syntastic) doesn't do well with gb. for this, we prebuild vendor packages second time, so go build also work correctly

cat manifest  | grep importpath | tr '"' '\n' | grep github | xargs -n1 go get -u
