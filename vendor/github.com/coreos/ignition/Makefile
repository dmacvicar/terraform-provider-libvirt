
.PHONY: vendor
vendor:
	@glide update --strip-vendor
	# TODO: Need --keep because update-ssh-keys uses symlinks within its package
	@glide-vc --use-lock-file --no-tests --only-code --keep '**/authorized_keys_d*'
