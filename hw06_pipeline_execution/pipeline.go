package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	current := in

	for _, stage := range stages {
		stageOutput := stage(current)

		// Обертка
		outWithDone := make(Bi)

		go func() {
			defer close(outWithDone)

			for {
				select {
				case v, ok := <-stageOutput:
					if !ok {
						return
					}

					select {
					case outWithDone <- v:
					case <-done:
						return
					}

				case <-done:
					return
				}
			}
		}()

		current = outWithDone
	}

	return current
}
