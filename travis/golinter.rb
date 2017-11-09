#! /usr/bin/ruby

require 'English'

# This list contain all files that are golint correct. (in master branch)
GOLINT_FILES_OK = %w[domain_def_test.go domain_def.go
                     resource_libvirt_domain_netiface.go
                     resource_libvirt_domain_console.go
                     qemu_agent_test.go utils_domain_def.go
                     libvirt_network_mock_test.go
                     utils_volume_test.go provider_test.go
                     utils_net_test.go network_def_test.go
                     utils_libvirt.go
                     pool_sync_test.go
                     volume_def_test.go utils_libvirt_test.go
                     disk_def_test.go
                     network_interface_def.go
                     libvirt_domain_mock_test.go ].freeze

# FIXME: blacklisted files need to go on GOLINT_FILES_OK once the lint are fixed
# The blacklist is usefull for catch golint error on new files
GOLINT_FILES_BLACKLISTED = %w[stream.go resource_libvirt_network.go
                              resource_libvirt_volume.go
                              resource_libvirt_volume_test.go
                              pool_sync.go
                              resource_libvirt_coreos_ignition_test.go
                              libvirt_interfaces.go
                              utils_test.go network_def.go disk_def.go
                              resource_libvirt_domain.go utils.go
                              provider.go resource_cloud_init.go
                              coreos_ignition_def.go utils_net.go qemu_agent.go
                              config.go resource_libvirt_coreos_ignition.go
                              volume_def.go cloudinit_def_test.go
                              cloudinit_def.go
                              utils_volume.go
                              resource_libvirt_domain_test.go ].freeze

# perform validation of the 2 lists
class Basic
  # scripts in blacklist are not cleaned to golint.
  # (remove them once you fix one)
  # this avoid that you forget
  def self.check_blacklist_always_fail
    GOLINT_FILES_BLACKLISTED.each do |f|
      `golint -set_exit_status #{f}`
      raise "GOFILE SHOULD FAIL! #{f}" if $CHILD_STATUS.exitstatus.zero?
    end
  end

  # check duplicata on two lists:
  # a go file cannot be in both lists
  def self.check_duplicata
    raise 'DUPLICATAS found!!' unless (GOLINT_FILES_BLACKLISTED &
                                                        GOLINT_FILES_OK).empty?
  end
end

def error_linter(file)
  puts '+' * 30
  puts "GOLINT FAILED for #{file}"
  puts '+' * 30
  exit 1
end

# Main tests

Dir.chdir 'libvirt/'
Basic.check_blacklist_always_fail
Basic.check_duplicata

Dir.foreach('.') do |file|
  next if (file == '.') || (file == '..')
  next if GOLINT_FILES_BLACKLISTED.include? file
  puts "Running golint for file #{file}"
  # Debug lint error for fixing them
  puts `golint -set_exit_status #{file}`
  error_linter(file) if $CHILD_STATUS.exitstatus.nonzero?
  puts "GOLINT OK! for #{file} !!!"
  puts '*' * 50
end
