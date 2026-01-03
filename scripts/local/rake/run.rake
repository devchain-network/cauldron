# frozen_string_literal: true

require 'English'

namespace :run do

  desc 'run webhookserver'
  task :webhookserver do
    system %{ go run -race cmd/webhookserver/main.go }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  namespace :kafka do
    namespace :github do

      desc 'run kafka github consumer'
      task :consumer do
        system %{ go run -race cmd/githubconsumer/main.go }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end

      desc 'run kafka github consumer group'
      task :consumer_group do
        system %{ go run -race cmd/githubconsumergroup/main.go }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end

    end

    namespace :gitlab do

      desc 'run kafka gitlab consumer'
      task :consumer do
        system %{ go run -race cmd/gitlabconsumer/main.go }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end

      desc 'run kafka gitlab consumer group'
      task :consumer_group do
        system %{ go run -race cmd/gitlabconsumergroup/main.go }
        exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
      rescue Interrupt
        exit(0)
      end

    end
  end
end
