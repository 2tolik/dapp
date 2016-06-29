module Dapp
  module Build
    module Stage
      class Source1Archive < SourceBase
        def initialize(application, next_stage)
          @prev_stage = InfraInstall.new(application, self)
          super
        end

        def prev_source_stage
          nil
        end

        def next_source_stage
          next_stage
        end

        def container_archive_path(git_artifact)
          application.container_build_path git_artifact.filename '.tar.gz'
        end

        protected

        def apply_command_method
          :archive_apply_command
        end
      end # Source1Archive
    end # Stage
  end # Build
end # Dapp
