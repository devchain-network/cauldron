# frozen_string_literal: true

require 'English'

namespace :run do

  desc 'run webhookserver'
  task :webhookserver do
    run = %{ go run -race cmd/webhookserver/main.go }
    pid = Process.spawn(run)
    Process.wait(pid)
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    Process.getpgid(pid)
    Process.kill('KILL', pid)
    exit(0)
  end

  namespace :kafka do
    namespace :github do

      desc 'run kafka github consumer'
      task :consumer do
        run = %{ go run -race cmd/githubconsumer/main.go }
        pid = Process.spawn(run)
        Process.wait(pid)
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        Process.getpgid(pid)
        Process.kill('KILL', pid)
        exit(0)
      end

      desc 'run kafka github consumer group'
      task :consumer_group do
        run = %{ go run -race cmd/githubconsumergroup/main.go }
        pid = Process.spawn(run)
        Process.wait(pid)
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        Process.getpgid(pid)
        Process.kill('KILL', pid)
        exit(0)
      end

    end
  end
end
